package saver

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/textproto"
	"path"
	"regexp"
	"time"

	".."
	"../../../RFC2047"
	"../../models"
	"../parser"
	"../storage"
)

// 当解析失败之后，但是也需要保存邮件的信息，此时就
// 尽可能从邮件头里面得到信息，然后存储到数据库里面去
func EmailSaveFallback(data []byte, uidl, message string, config *models.ServerConfig) {
	var idx = bytes.Index(data, []byte("\n\n"))

	if idx == -1 {
		idx = len(data)
	}

	var reader = textproto.NewReader(bufio.NewReader(bytes.NewReader(data[0:idx])))
	var header, err = reader.ReadMIMEHeader()
	if header == nil || len(header) <= 0 {
		// 这里不判断err，而是判断header，只要header可以用，就继续往下面执行
		log.Println(err)
		return
	}

	date, err := util.ParseDate(header.Get("Date"))
	if err != nil {
		log.Println(uidl, header.Get("Date"), err)
		date = time.Now()
	}

	EmailSave(&models.Email{
		Subject: parser.FixSubject(RFC2047.Decode(header.Get("Subject"))),
		Date:    date,
		Uidl:    uidl,
		From:    header.Get("From"),
		Cc:      header.Get("Cc"),
		Bcc:     header.Get("Bcc"),
		To:      header.Get("To"),
		ReplyTo: header.Get("Reply-To"),
		Status:  3,
		Message: message,
		IsRead:  1,
	}, config)
}

// 把邮件的保存到数据库，包括相关的Tags信息
func EmailSave(email *models.Email, config *models.ServerConfig) {
	var err error

	email.Id, err = config.Ormer.Insert(email)
	if err != nil {
		log.Println(err)
		return
	}

	// 看看是否有Tags的信息，如果有的话，需要更新Tags
	if email.Tags != nil && len(email.Tags) > 0 {
		// 初始化Tag的信息，有的话，得到Id，没有的话，插入之后也得到了Id
		for _, tag := range email.Tags {
			_, tag.Id, _ = config.Ormer.ReadOrCreate(tag, "Name")
		}

		var m2m = config.Ormer.QueryM2M(email, "Tags")
		_, err = m2m.Add(email.Tags)
		if err != nil {
			log.Println("[ ADD TAG]", email.Uidl, err)
		}
	}
}

// 把邮件相关的资源保存下来，以本地文件或者网盘的文件的形式
func EmailResourceSave(email *models.Email, config *models.ServerConfig) {
	re := regexp.MustCompile(`src="cid:([^"]+)"`)
	sm := re.FindAllSubmatch([]byte(email.Message), -1)

	// 先保存cid
	for _, match := range sm {
		var key = string(match[1])
		if value, ok := email.ResourceBundle[key]; ok {
			// 如果存在的话，那么这个文件需要写入cid目录
			var dst = path.Join(config.BaseDir, "downloads", email.Uidl, "cid", key)
			var data = value.Body
			storage.NewDiskStorage(dst, data, 0644).Save()

			// 写完之后删除，最后剩下的就放到att目录即可
			delete(email.ResourceBundle, key)
		}
	}

	// 再保存att
	if len(email.ResourceBundle) <= 0 {
		return
	}

	for _, value := range email.ResourceBundle {
		var filename string
		if value.Name != "" {
			filename = value.Name
		} else if value.ContentId != "" {
			filename = value.ContentId
		} else {
			continue
		}

		var dst = path.Join(config.BaseDir, "downloads", email.Uidl, "att", filename)
		var data = value.Body

		// TODO(user) 应该通过 chan 传递数据过去，而不是每次启动一个新的 goroutine
		go storage.NewDiskStorage(dst, data, 0644).Save()

		// dst需要重新计算
		var token = config.Service.Netdisk.AccessToken
		if token != "" {
			var uidl = util.StripInvalidCharacter(email.Uidl)
			var name = util.StripInvalidCharacter(filename)
			if uidl != "" && name != "" {
				// 一般不会超过1000个字节，所以不考虑超长的情况了
				var dst = fmt.Sprintf("/apps/dropbox/%s/%s/%s/%s",
					config.Pop3.Domain, config.Pop3.Username, uidl, name)
				if len([]byte(dst)) > 1000 {
					log.Println(dst, "was too long")
					continue
				}

				// TODO(user) 应该通过 chan 传递数据过去，而不是每次启动一个新的 goroutine
				go storage.NewNetdiskStorage(token, dst, data).Save()
			}
		}
	}
}