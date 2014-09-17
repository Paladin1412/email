package base

const (
	kActionMoveMessage              = "MoveMessage"
	kActionCopyMessage              = "CopyMessage"
	kActionDeleteMessage            = "DeleteMessage"
	kActionLabel                    = "Label"
	kActionMarkAsRead               = "MarkAsRead"
	kActionChangeStatus             = "ChangeStatus"
	kActionToDo                     = "ToDo"
	kActionReply                    = "Reply"
	kActionForward                  = "Forward"
	kActionSaveAttachments          = "SaveAttachments"
	kActionDoNotNotify              = "DoNotNotify"
	kSubject                        = "Subject"
	kFrom                           = "From"
	kTo                             = "To"
	kCc                             = "Cc"
	kBcc                            = "Bcc"
	kReplyTo                        = "Reply-To"
	kDate                           = "Date"
	kSentTo                         = "SentTo"
	kBody                           = "Body"
	kSubjectOrBody                  = "SubjectOrBody"
	kSize                           = "Size"
	kAttachments                    = "Attachments"
	kContentType                    = "Content-Type"
	kQuotedPrintable                = "quoted-printable"
	kBase64                         = "base64"
	kContentId                      = "Content-ID"
	kMessageId                      = "Message-ID"
	kReferences                     = "References"
	kInReplyTo                      = "In-Reply-To"
	kContentDisposition             = "Content-Disposition"
	kContentTransferEncoding        = "Content-Transfer-Encoding"
	kTimeLayout                     = "Mon, 2 Jan 2006 15:04:05 -0700"
	kDefaultInterval                = 60
	kDefaultDownloadDir             = "downloads"
	kDefaultRawDir                  = "raw"
	kDefaultDbName                  = "foo.db"
	kDefaultPop3Port                = 995
	kDefautlSmtpPort                = 25
	kDefaultContentType             = "text/html; charset=\"utf-8\""
	kDefaultContentTransferEncoding = "base64"
	kOpIs                           = "Is"
	kOpContains                     = "Contains"
	kOpExists                       = "Exists"
	KOpOnlyMe                       = "Only Me"
	kOpMe                           = "Me"
	kOpNotMe                        = "Not Me"
	kOpCcMe                         = "Cc Me"
	kOpToOrCcMe                     = "To or Cc Me"
	kOpHasAttachments               = "HasAttachments"
	kOpRange                        = "Range"
	kMatchAll                       = "All"
	kMatchAny                       = "Any"
	kLogFormat                      = "%{color}%{time:15:04:05.000000} %{level:.4s} %{id:03x}%{color:reset} %{message}"
)
