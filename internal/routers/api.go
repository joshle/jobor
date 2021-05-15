package routers

import (
	"github.com/gin-gonic/gin"
	"jobor/internal/app/auth"
	"jobor/internal/app/jobor/dispatcher"
	joborlog "jobor/internal/app/jobor/log"
	joborWorker "jobor/internal/app/jobor/worker"
	"jobor/internal/app/sys/permission"
	role2 "jobor/internal/app/sys/role"
	"jobor/internal/app/sys/server"
	user2 "jobor/internal/app/sys/user"
	"jobor/internal/app/sys/usergroup"
	"jobor/internal/models/db"
	"jobor/pkg/logger"
	"jobor/pkg/utils"
	"net"
)

// RegisterRouter 注册/api路由
func RegisterRouter(engine *gin.Engine) error {
	v1 := engine.Group("/api/v1")
	{
		v1.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "test 获取成功.",
				"status": "success",
				"data": []string{},
				//"error": err,
			})
		})
	}

	engine.GET("/ping", Ping)
	engine.GET("/gin/routes", Query)

	engine.POST("/api/v1/login", auth.LoginAuth)
	engine.POST("/api/v1/refresh-token", auth.RefreshToken)
	engine.POST("/api/login", auth.LoginAuth)
	engine.GET("/adders", auth.VerifyCode)
	engine.GET("/api/v1/user-info", user2.GetUserInfo)

	{
		// ################ system ###################
		sys := v1.Group("/sys")
		{
			var user = user2.NewService(db.DB)
			sys.GET("/user/:id", user.Get)
			sys.GET("/users", user.GetAll)
			sys.GET("/user", user.Query)
			sys.POST("/user", user.Create)
			sys.PUT("/user/:id", user.Update)
			sys.DELETE("/user", user.Delete)

			var userGroup = usergroup.NewService(db.DB)
			sys.GET("/usergroups", userGroup.GetAll)
			sys.GET("/usergroup", userGroup.Query)
			sys.POST("/usergroup", userGroup.Create)
			sys.PUT("/usergroup", userGroup.Update)
			sys.DELETE("/usergroup", userGroup.Delete)

			var role = role2.NewService(db.DB)
			sys.GET("/roles", role.GetAll)
			sys.GET("/role", role.Query)
			sys.POST("/role", role.Create)
			sys.PUT("/role", role.Update)
			sys.DELETE("/role", role.Delete)

			var permissionM = permission.NewService(db.DB)
			sys.GET("/permissions", permissionM.Query)
			sys.GET("/permission", permissionM.Query)
			sys.POST("/permission", permissionM.Create)
			sys.PUT("/permission/:id", permissionM.Update)
			sys.DELETE("/permission", permissionM.Delete)


			var state = server.NewService(db.DB)
			sys.GET("/state", state.Get)
		}


		// ################ Jobor ###################
		joborApp := v1.Group("/jobor")
		{
			var joborTask = dispatcher.NewService(db.DB)
			joborApp.GET("/dashboard", joborTask.Dashboard)
			joborApp.GET("/task", joborTask.Query)
			joborApp.POST("/task", joborTask.Create)
			joborApp.PUT("/task/:id", joborTask.Update)
			joborApp.POST("/task/:id/run", joborTask.RunTask)
			joborApp.PUT("/task/:id/:status", joborTask.RunOrStopStatus)
			joborApp.DELETE("/task/:id", joborTask.Delete)

			var iJoborWorker = joborWorker.NewService(db.DB)
			joborApp.GET("/worker", iJoborWorker.Query)
			joborApp.PUT("/worker/:id", iJoborWorker.Update)
			joborApp.DELETE("/worker/:id", iJoborWorker.Delete)
			var joborTaskLog = joborlog.NewService(db.DB)
			joborApp.GET("/log", joborTaskLog.Query)
			joborApp.POST("/log/:id/abort", joborTaskLog.Abort)
		}


	}

	return nil
}

var adminRoles = []string{"administrators", "admin", "root"}
func callHandlerPermission(roles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles = utils.Union(roles, adminRoles)
		//fmt.Println("roles:", roles)
		userRoles := []string{"bb", "cc", "devops"}
		intersects := utils.Intersect(roles, userRoles)
		//fmt.Println("intersect:", intersects, len(intersects))

		/**/
		if len(intersects) > 0 {
			c.Next()
			return
		}else {
			c.JSON(200, gin.H{
				"status": "error",
				"message": "没有权限",
				"code": 401,
			})
			c.Abort()
			return
		}

	}
}


func Ping(c *gin.Context)  {
	logger.Infof("from client %s ping", c.ClientIP())
	ips,_:= ServerIPv4s()
	c.JSON(200, gin.H{
		"app": "jobor.inc",
		"message": "pong",
		"status": "success",
		"clientIP": c.ClientIP(),
		"serverIPs": ips,
	})
}

// ServerIPv4s LocalIPs return all non-loopback IPv4 addresses
func ServerIPv4s() ([]string, error) {
	var ips []string
	adders, err := net.InterfaceAddrs()

	if err != nil {
		return ips, err
	}

	for _, a := range adders {
		if inet, ok := a.(*net.IPNet); ok && !inet.IP.IsLoopback() && inet.IP.To4() != nil {
			ips = append(ips, inet.IP.String())
		}
	}
	return ips, nil
}