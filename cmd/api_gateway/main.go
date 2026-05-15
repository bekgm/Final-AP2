package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	pb "github.com/bekgm/Final-AP2/pkg/messaging"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	_ = godotenv.Load()

	grpcTarget := os.Getenv("GRPC_TARGET")
	if grpcTarget == "" {
		grpcTarget = "localhost:50051"
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(grpcTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewMessagingServiceClient(conn)

	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		api.POST("/messages", func(c *gin.Context) {
			var req struct {
				ReceiverID string `json:"receiver_id" binding:"required"`
				ProjectID  string `json:"project_id"`
				Content    string `json:"content" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// In a real app, you would extract the user ID from the JWT token.
			// Here we take it from a custom header for simplicity.
			senderID := c.GetHeader("X-User-ID")
			if senderID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header is required"})
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			res, err := client.SendMessage(ctx, &pb.SendMessageRequest{
				SenderId:   senderID,
				ReceiverId: req.ReceiverID,
				ProjectId:  req.ProjectID,
				Content:    req.Content,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, res.Message)
		})

		api.GET("/messages", func(c *gin.Context) {
			userID1 := c.GetHeader("X-User-ID")
			if userID1 == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header is required"})
				return
			}

			userID2 := c.Query("user_id")
			if userID2 == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter is required"})
				return
			}

			projectID := c.Query("project_id")
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			res, err := client.GetMessages(ctx, &pb.GetMessagesRequest{
				UserId_1:  userID1,
				UserId_2:  userID2,
				ProjectId: projectID,
				Limit:     int32(limit),
				Offset:    int32(offset),
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, res.Messages)
		})

		api.GET("/dialogs", func(c *gin.Context) {
			userID := c.GetHeader("X-User-ID")
			if userID == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID header is required"})
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			res, err := client.GetDialogs(ctx, &pb.GetDialogsRequest{
				UserId: userID,
			})

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, res.Dialogs)
		})
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to run api gateway: %v", err)
	}
}
