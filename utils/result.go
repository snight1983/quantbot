package utils

func init() {

}

type Return struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data"`
}

func GetReturn(code int, msg interface{}, res interface{}) *Return {
	return &Return{
		Code: code,
		Msg:  msg,
		Data: res,
	}
}
