package entity

import "time"

// TimeResponse 定义了包含状态码、消息和当前时间的数据响应结构
type TimeResponse struct {
    Code int `json:"code"`
    Msg  string `json:"msg"`
    Data struct {
        Time string `json:"time"`
    } `json:"data"`
}

// NewTimeResponse 创建一个新的TimeResponse实例，带有给定的状态码、消息以及当前格式化后的时间
func NewTimeResponse(code int, msg string) TimeResponse {
    currentTime := time.Now().Format("2006-01-02 15:04:05")
    return TimeResponse{
        Code: code,
        Msg:  msg,
        Data: struct {
            Time string `json:"time"`
        }{Time: currentTime},
    }
}