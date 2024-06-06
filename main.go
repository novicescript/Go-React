package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	Id primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool `json:"completed"`
	Body string `json:"body"`
}

var collection *mongo.Collection

func main() {
	fmt.Println("Hello world!")

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("ERROR IN LOADING ENV")
		}
	}

	port := os.Getenv("PORT")
	MONGODB_URI := os.Getenv("MONGO_URI")
	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("CONNECTED TO MONGODB")

	collection = client.Database("go-react").Collection("todos")

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://timelapse.onrender.com",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)


	if port == "" {
		port = "5000"
	}

	log.Fatal(app.Listen("0.0.0.0:"+port))
}


func getTodos(c *fiber.Ctx) error {
	var todos []Todo 
	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err 
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo 
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}

	return c.Status(200).JSON(fiber.Map{
		"status": "success", 
		"message": "All Your Todos", 
		"total": len(todos), 
		"todos": todos,
	})
}


func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"message": "Todo body can not be empty!"})
	}

	// existBody, err := collection.FindOne(context.Background(), bson.M{})

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	todo.Id = insertResult.InsertedID.(primitive.ObjectID)
	

	return c.Status(201).JSON(fiber.Map{
		"status": "success", 
		"message": "Added new Todo", 
		"todo": todo,
	})
}


func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Todo not found with given Id"})
	}


	filter := bson.M{"_id":  objectId}
	query := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, query)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"message": "Todo updated"})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid ID."})
	}

	filter := bson.M{"_id":objectId}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"status": "success", "message": "Todo Deleted"})

}
