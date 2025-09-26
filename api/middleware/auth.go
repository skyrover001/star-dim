package middleware

//func JWTAuth(userDB *models.UserDB) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		authHeader := c.GetHeader("Authorization")
//		if authHeader == "" {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要Authorization头"})
//			c.Abort()
//			return
//		}
//
//		parts := strings.Split(authHeader, " ")
//		if len(parts) != 2 || parts[0] != "Bearer" {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization头格式必须为Bearer <token>"})
//			c.Abort()
//			return
//		}
//
//		tokenString := parts[1]
//		claims, err := utils.ValidateToken(tokenString)
//		if err != nil {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的令牌"})
//			c.Abort()
//			return
//		}
//
//		user, err := userDB.GetUserByID(claims.UserID)
//		if err != nil {
//			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被删除"})
//			c.Abort()
//			return
//		}
//
//		c.Set("user_id", user.ID)
//		c.Set("username", user.Username)
//		c.Set("user_role", user.Role)
//		c.Set("user", user)
//
//		c.Next()
//	}
//}
