package thread

import (
	"regexp"
)

type ContainerMap map[string]*Container

type Thread struct {
	idTable      ContainerMap
	subjectTable ContainerMap
}

func (t *Thread) createIdTable(messages []*Message) {
	// 初始化
	t.idTable = make(ContainerMap)

	for _, m := range messages {
		parentContainer, ok := t.idTable[m.Id]
		if ok {
			// 如果 id_table 包含这个 message-id 的话，检查一下是否是
			// empty container
			if parentContainer.IsEmpty() {
				parentContainer.message = m
			}
		} else {
			// id_table 不包含这个 message-id，那么创建一个新的
			parentContainer = &Container{message: m}
			t.idTable[m.Id] = parentContainer
		}

		// Link the References field's Containers together in the order
		// implied by the References header.
		// For each element in the message's References field:
		var prev *Container
		for _, ref := range m.References {
			container, ok := t.idTable[ref]
			if !ok {
				// If there's one in id_table use that;
				// Otherwise, make (and index) one with a null Message.
				container = &Container{message: nil}
				t.idTable[ref] = container
			}

			if prev != nil &&
				// If they are already linked, don't change the existing links.
				container.parent == nil &&
				// Do not add a link if adding that link would introduce a loop:
				// that is, before asserting A->B, search down the children of B
				// to see if A is reachable, and also search down the children of
				// A to see if B is reachable. If either is already reachable as
				// a child of the other, don't add the link.
				!container.HasDescendant(prev) {
				prev.AddChild(container)
			}

			prev = container
		}

		// Set the parent of this message to be the last element in References.
		// Note that this message may have a parent already: this can happen because we saw this ID
		// in a References field, and presumed a parent based on the other entries in that field.
		// Now that we have the actual message, we can be more definitive, so throw away the old parent
		// and use this new one. Find this Container in the parent's children list, and unlink it.

		// Note that this could cause this message to now have no parent, if it has no references
		// field, but some message referred to it as the non-first element of its references.
		// (Which would have been some kind of lie...)

		// Note that at all times, the various ``parent'' and ``child'' fields must be
		// kept inter-consistent.
		if prev != nil && !parentContainer.HasDescendant(prev) {
			prev.AddChild(parentContainer)
		}
	}
}

func (t *Thread) GroupBySubject(roots *Container) ContainerMap {
	// 初始化
	t.subjectTable = make(ContainerMap)

	// 先构造一个初级的subjectTable
	for _, container := range roots.children {
		var this *Container
		if container.IsEmpty() {
			if len(container.children) > 0 {
				this = container.children[0]
			}
		} else {
			this = container
		}

		if this == nil || this.IsEmpty() {
			continue
		}

		subject := normalizeSubject(this.message.Subject)
		if subject == "" {
			continue
		}

		if old, ok := t.subjectTable[subject]; !ok {
			// There is no container in the table with this subject
			t.subjectTable[subject] = this
		} else {
			// 1. This one is an empty container and the old one is not:
			// the empty one is more interesting as a root, so put it in the table instead.
			//
			// 2. The container in the table has a ``Re:'' version of this subject,
			// and this container has a non-``Re:'' version of this subject.
			// The non-re version is the more interesting of the two.
			if !old.IsEmpty() {
				if this.IsEmpty() {
					t.subjectTable[subject] = this
				} else if isReplyOrForward(old.message.Subject) &&
					!isReplyOrForward(this.message.Subject) {
					t.subjectTable[subject] = this
				}
			}

		}
	}

	for _, container := range roots.children {
		subject := normalizeSubject(container.GetSubject())

		c := t.subjectTable[subject]
		if c == nil || c == container {
			continue
		}

		if c.IsEmpty() && container.IsEmpty() {
			// If both are dummies, append one's children to the other,
			// and remove the now-empty container.
			for _, child := range container.children {
				c.AddChild(child)
			}
			container.parent.RemoveChild(container)
		} else if c.IsEmpty() && !container.IsEmpty() {
			// If one container is a empty and the other is not,
			// make the non-empty one be a child of the empty,
			// and a sibling of the other ``real'' messages with
			// the same subject (the empty's children.)
			c.AddChild(container)
		} else if !isReplyOrForward(c.message.Subject) &&
			// TODO(user) 为啥有这个判断了呢?
			!container.IsEmpty() &&
			isReplyOrForward(container.message.Subject) {
			// If that container is a non-empty, and that message's
			// subject does not begin with ``Re:'', but this message's
			// subject does, then make this be a child of the other.
			c.AddChild(container)
		} else {
			nc := &Container{message: nil}
			nc.AddChild(c)
			nc.AddChild(container)
			t.subjectTable[subject] = nc
		}
	}

	return t.subjectTable
}

func (t *Thread) GetRoots() *Container {
	roots := &Container{message: nil}

	for _, child := range t.idTable {
		if child.parent == nil {
			roots.AddChild(child)
		}
	}

	roots.PruneEmpties()

	return roots
}

func normalizeSubject(subject string) string {
	re := regexp.MustCompile(`(?i)((Re|Fwd|回复|答复)(\[[\d+]\])?[:：](\s*)?)*(.*)`)
	ss := re.FindStringSubmatch(subject)
	return ss[5]
}

func isReplyOrForward(subject string) bool {
	re := regexp.MustCompile(`^(?i)(Re|Fwd|回复|答复)`)
	return re.MatchString(subject)
}

func NewThread(messages []*Message) *Thread {
	thread := &Thread{}
	thread.createIdTable(messages)

	return thread
}
