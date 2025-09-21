package router

import (
	"github.com/gin-gonic/gin"
	"im.api/internal/handler"
	"im.api/internal/middleware"
	"im.api/internal/ws"
)

func InitRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", handler.PingHandler)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/register", handler.RegisterHandler)
		v1.POST("/login", handler.LoginHandler)
		v1.GET("/captcha", handler.GetCaptchaHandler)
		v1.GET("/ws", middleware.WSJWTAuthMiddleware(), ws.WsHandler)
	}
	file := v1.Group("/file", middleware.JWTAuthMiddleware())
	{
		file.POST("/check-hash", handler.CheckFileHashHandler)
		file.POST("/upload", handler.UploadFileHandler)
		file.GET("/", handler.GetFileHandler)
	}
	user := v1.Group("/user", middleware.JWTAuthMiddleware())
	{
		user.GET("/info", handler.UserInfoHandler)
		user.PUT("/info", handler.UpdateUserInfoHandler)
		user.POST("/avatar", handler.UploadUserAvatarHandler)
		user.PUT("/password", handler.ChangePasswordHandler)
		user.GET("/devices", handler.GetDevicesHandler)
		user.DELETE("/devices", handler.LogoutDeviceHandler)
		user.GET("/search", handler.SearchUserHandler)
		user.GET("/offlinemsg", handler.OfflineMsgHandler)
		user.GET("/msg", handler.MsgHandler)
		user.POST("/msg", handler.SendMsgHandler)
	}

	// Friend API routes
	friend := v1.Group("/friend", middleware.JWTAuthMiddleware())
	{
		friend.POST("/request", handler.SendFriendRequestHandler)  // 发送好友请求
		friend.PUT("/request", handler.HandleFriendRequestHandler) // 处理好友请求
		friend.GET("/requests", handler.GetFriendRequestsHandler)  // 获取好友请求列表
		friend.GET("/list", handler.GetFriendListHandler)          // 获取好友列表
		friend.DELETE("/", handler.DeleteFriendHandler)            // 删除好友
		friend.GET("/info", handler.GetFriendInfoHandler)          //获取好友详细信息

	}

	group := v1.Group("/group", middleware.JWTAuthMiddleware())
	{
		group.GET("/search", handler.GetGroupByNumberHandler)            // 通过群号码查询群组
		group.POST("/", handler.CreateGroupHandler)                      // 创建群组
		group.GET("/list", handler.GetUserGroupsHandler)                 // 获取用户参与的群组列表
		group.POST("/join_request", handler.SendJoinGroupRequestHandler) // 发送加群信息
		member := group.Group("/member")
		{
			member.DELETE("/", handler.LeaveGroupHandler)    // 退出群组
			member.GET("/info", handler.GetGroupInfoHandler) // 获取群组详细信息
		}
		admin := group.Group("/admin", middleware.AdminAuthMiddleware())
		{
			admin.PUT("/", handler.UpdateGroupInfoHandler)                     // 更新群组信息
			admin.POST("/avatar", handler.UpdateGroupAvatarHandler)            // 更新群组头像
			admin.POST("/join_request", handler.HandleJoinGroupRequestHandler) // 处理入群信息
			admin.GET("/requests", handler.GetGroupsRequestsHandler)
			admin.DELETE("/member", handler.KickGroupMemberHandler) // 踢出成员
			admin.DELETE("/", handler.DissolveGroupHandler)         // 解散群组
		}

	}
	static := r.Group("/static", middleware.JWTAuthMiddleware())
	static.Static("/avatar", "static/avatar")

	return r
}
