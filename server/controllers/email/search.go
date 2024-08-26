package email

import (
	"database/sql"
	"encoding/json"
	"github.com/Jinnrry/pmail/config"
	"github.com/Jinnrry/pmail/db"
	"github.com/Jinnrry/pmail/dto/response"
	"github.com/Jinnrry/pmail/utils/array"
	"io"
	"net/http"
	"strings"
	"time"
	"xorm.io/builder"
)

type Email struct {
	Id           int            `json:"id"`
	Type         int8           `json:"type"`
	Subject      string         `json:"subject"`
	ReplyTo      string         `json:"reply_to"`
	FromName     string         `json:"from_name"`
	FromAddress  string         `json:"from_address"`
	To           string         `json:"to"`
	Bcc          string         `json:"bcc"`
	Cc           string         `json:"cc"`
	Text         sql.NullString `json:"text"`
	Html         sql.NullString `json:"html"`
	Sender       string         `json:"sender"`
	Attachments  string         `json:"attachments"`
	SPFCheck     int8           `json:"spf_check"`
	DKIMCheck    int8           `json:"dkim_check"`
	Status       int8           `json:"status"`
	CronSendTime time.Time      `json:"cron_send_time"`
	UpdateTime   time.Time      `json:"update_time"`
	SendUserID   int            `json:"send_user_id"`
	Size         int            `json:"size"`
	Error        sql.NullString `json:"error"`
	SendDate     time.Time      `json:"send_date"`
	CreateTime   time.Time      `json:"create_time"`
}

type searchRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
}

func Search(w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)

	if err != nil {
		//log.WithContext(ctx).Errorf("%+v", err)
		response.NewErrorResponse(response.ParamsError, "params error", err.Error()).FPrint(w)
		return
	}

	//log.WithContext(ctx).Infof("查询邮件")
	var reqData searchRequest

	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		//log.WithContext(ctx).Errorf("%+v", err)
		response.NewErrorResponse(response.ParamsError, "params error", err.Error()).FPrint(w)
		return
	}

	if reqData.From == "" {
		response.NewErrorResponse(response.ParamsError, "发件人必填", "发件人必填").FPrint(w)
		return
	}

	if reqData.To == "" {
		response.NewErrorResponse(response.ParamsError, "收件人必填", "收件人必填").FPrint(w)
		return
	}

	infos := strings.Split(reqData.To, "@")
	if len(infos) != 2 || !array.InArray(infos[1], config.Instance.Domains) {
		response.NewErrorResponse(response.ParamsError, "params error", "").FPrint(w)
		return
	}

	if reqData.Subject == "" {
		response.NewErrorResponse(response.ParamsError, "邮件标题必填", "邮件标题必填").FPrint(w)
		return
	}

	emails, err := searchEmails(reqData.From, reqData.To)
	if err != nil {
		//log.WithContext(ctx).Errorf("搜索邮件错误: %+v", err)
		response.NewErrorResponse(response.ServerError, "搜索失败", err.Error()).FPrint(w)
		return
	}

	resp := struct {
		Emails []Email `json:"emails"`
	}{
		Emails: emails,
	}

	response.NewSuccessResponse(resp).FPrint(w)

}

func searchEmails(fromAddress string, toAddress string) ([]Email, error) {
	query := db.Instance.Table("email").Select("*")
	cond := builder.NewCond()

	if fromAddress != "" {
		cond = cond.And(builder.Eq{"from_address": fromAddress})
	}

	if toAddress != "" {
		cond = cond.And(builder.Expr(`"to"::jsonb @> jsonb_build_array(jsonb_build_object('EmailAddress', $2::text))`, toAddress))
	}

	query = query.Where(cond).Desc("id")

	var emails []Email
	err := query.Find(&emails)

	if err != nil {
		return nil, err
	}

	return emails, nil
}
