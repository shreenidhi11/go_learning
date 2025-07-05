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
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Completed bool `json:"completed"`
	Body string `json:"body"`
}

// create a pointer to the mongo collection
var collection *mongo.Collection

func main()  {
	fmt.Println("Hello Worlds")

	// load the env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	// load the url of the mongodb from the .env
	MONGO_URI := os.Getenv("MONGO_URI")
	// create client options that configures how to connect to MongoDB
	clientOptions := options.Client().ApplyURI(MONGO_URI)

	// connecting to mongodb using the client options
	client,err := mongo.Connect(context.Background(),clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// To Verify that the database is reachable
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// final message
	fmt.Println("Connected to MongoDB Atlas")

	// collection now points to todos collection in atlas
	collection = client.Database("golang_db").Collection("todos")

	// create the fiber object
	app := fiber.New()

	// Enable CORS for all origins
	app.Use(cors.New())

	// creating all the endpoints
	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodos)
	app.Patch("/api/todos/:id", updateTodos)
	app.Delete("/api/todos/:id", deleteTodos)

	// set the port
	port := os.Getenv("PORT")
	if port == ""{
		port = "5000"
	}


	log.Fatal(app.Listen("0.0.0.0:" + port))

}

func getTodos(c *fiber.Ctx) error {
	// creating a dynamic array of todos to store the todos from mongodb
	var todos []Todo

	// get all the todos from mongodb and store it in a cursor
	// here bson.M means nothing to match
	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err
	}

	// we close the cursor after the getTodos method execution 
	defer cursor.Close(context.Background())

	// looping the cursor
	for cursor.Next(context.Background()) {
		// create a todo varaible of type todo and check if the mongodb return document matches
		// the struct 
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		// append if correct
		todos = append(todos, todo)
	}

	// return all todos
	return c.JSON(todos)
}

func createTodos(c *fiber.Ctx) error {
	todo := new(Todo)
	// {id:0,completed:false,body:""}

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	// Had to change the struct definition to support the unique ID generator from mongoDB
	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func updateTodos(c *fiber.Ctx) error {
	id := c.Params("id")
	// Converts string ID from the URL to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	// performing the filter and update of the required document
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")

	// Converts string ID from the URL to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	// performing the filter and update of the required document
	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})
}