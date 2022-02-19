package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fufuok/xy-data-router/common"
	"github.com/fufuok/xy-data-router/internal/json"
)

var apiSuccessNil = json.MustJSON(common.APISuccessNil())

// APIException 通用异常处理
func APIException(c *fiber.Ctx, code int, msg string) error {
	if msg == "" {
		msg = "错误的请求"
	}
	return c.Status(code).JSON(common.APIFailureData(msg))
}

// APIFailure 返回失败, 状态码: 200
func APIFailure(c *fiber.Ctx, msg string) error {
	return APIException(c, fiber.StatusOK, msg)
}

// APISuccess 返回成功, 状态码: 200
func APISuccess(c *fiber.Ctx, data interface{}, count int) error {
	return c.JSON(common.APISuccessData(data, count))
}

// APISuccessBytes 返回成功, JSON 字节数据, 状态码: 200
func APISuccessBytes(c *fiber.Ctx, data []byte, count int) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Send(common.APISuccessBytes(data, count))
}

// APISuccessNil 返回成功, 无数据, 状态码: 200
func APISuccessNil(c *fiber.Ctx) error {
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return c.Send(apiSuccessNil)
}

// TxtMsg 返回文本消息
func TxtMsg(c *fiber.Ctx, msg string) error {
	return c.SendString(msg)
}
