package schema

import (
	"net/http"
	"strconv"

	"github.com/gorilla/schema"
)

type MailListSchema struct {
	PageSize   int `schema:"pageSize"`
	PageNo     int `schema:"pageNo"`
	LabelId    int `schema:"label"`
	UnRead     int `schema:"unreadOnly"`
	IsDelete   int `schema:"is_delete"`
	IsStar     int `schema:"is_star"`
	IsSent     int `schema:"is_sent"`
	IsCalendar int `schema:"is_calendar"`
	SkipCount  int `schema:"-"`
}

func (this *MailListSchema) getSearchCriteria() string {
	var sql string

	if this.IsDelete == 1 {
		sql += "WHERE `is_delete` = 1 "
	} else if this.IsStar == 1 {
		sql += "WHERE `is_star` = 1 "
	} else {
		if this.LabelId > 0 {
			if this.LabelId != 2 && this.LabelId != 9 {
				sql += "WHERE `is_delete` != 1 AND `id` IN " +
					"(SELECT `mid` FROM `mail_tags` WHERE `tid` = " +
					strconv.Itoa(this.LabelId) + ") "
			} else {
				// Spam 和 监控邮件 默认设置 is_delete = 1，因此
				// 如果有 is_delete = 1在这种情况下，是过滤不出来任何东西的
				sql += "WHERE `id` IN " +
					"(SELECT `mid` FROM `mail_tags` WHERE `tid` = " +
					strconv.Itoa(this.LabelId) + ") "
			}
		} else {
			sql += "WHERE `is_delete` != 1 "
		}
	}

	if this.UnRead == 1 {
		// 如果有明确的标识说只看未读的邮件，才加上这个条件，否则返回未读和已读的
		sql += "AND `is_read` != 1 "
	}

	if this.IsStar != 1 {
		if this.IsSent == 1 {
			sql += "AND `is_sent` = 1 "
		} else {
			sql += "AND `is_sent` != 1 "
		}

		if this.IsCalendar == 1 {
			sql += "AND `is_calendar` = 1 "
		} else {
			sql += "AND `is_calendar` != 1 "
		}
	}

	return sql
}

func (this *MailListSchema) BuildListSql() string {
	// 准备sql
	sql := "SELECT " +
		"`id`, `uidl`, `from`, `to`, `cc`, `bcc`, " +
		"`reply_to`, `subject`, `date`, `is_read`, `is_star` " +
		"FROM mails "
	sql += this.getSearchCriteria()
	sql += "ORDER BY `date` DESC, `id` DESC LIMIT ?, ?"

	return sql
}

func (this *MailListSchema) BuildTotalSql() string {
	sql := "SELECT COUNT(*) FROM mails "

	sql += this.getSearchCriteria()

	return sql
}

func (this *MailListSchema) Init(r *http.Request) {
	r.ParseForm()

	schema.NewDecoder().Decode(this, r.PostForm)
	this.setDefault()
}

func (this *MailListSchema) setDefault() {
	if this.PageSize == 0 {
		this.PageSize = kDefaultPageSize
	}

	if this.PageNo == 0 {
		this.PageNo = kDefaultPageNo
	}

	if this.LabelId == 0 {
		this.LabelId = kDefaultLabelNo
	}

	this.SkipCount = (this.PageNo - 1) * this.PageSize
	if this.SkipCount < 0 {
		this.SkipCount = 0
	}
}
