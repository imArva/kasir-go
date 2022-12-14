package main

import (
	"database/sql"
	"embed"
	"log"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed views/*
var viewsfs embed.FS

var db *sql.DB

func main(){
	var err error
	db, err = sql.Open("sqlite3", "./kasir.sqlite3")
	if err != nil {
		log.Fatal("falide to connect to database")
	}
	defer db.Close()

	engine := html.NewFileSystem(http.FS(viewsfs), ".html")

	app := fiber.New(fiber.Config{Views: engine})

	app.Get("/", getIndex)
	app.Get("/additem", getAddItem)
	app.Get("/data", getData)
	app.Get("/edit", getEdit)
	app.Get("/update/:id", getUpdate)
	app.Get("/delete/:id", Delete)

	app.Post("/addprocess", postAddItem)
    app.Post("/editprocess", postEdit)
    app.Post("/postdata", postData)

	log.Fatal(app.Listen(":8080"))
}

func getIndex(c *fiber.Ctx) error {
	type Item struct{
		Id int `json:"id"`
		NamaItem string `json:"Nama_item"`
		HargaItem int `json:"Harga_item"`
	}
	
	rows, err := db.Query("SELECT id, nama_item, harga_item FROM items")
	if err != nil {
		c.SendString(err.Error())
	}
	defer rows.Close()

	items := []*Item{}
	var lastId int
	for rows.Next() {
		var buffer Item
		err := rows.Scan(&buffer.Id,&buffer.NamaItem,&buffer.HargaItem)
		if err != nil {
			return c.SendString(err.Error())
		}
		lastId = buffer.Id
		items = append(items, &buffer)
	}

	return c.Render("views/index", fiber.Map{
		"Items": &items,
		"LastId": lastId,
	})
}

func getAddItem(c *fiber.Ctx) error {
	return c.Render("views/item", fiber.Map{})
}

func getData(c *fiber.Ctx) error {
	type Item struct{
		Id int `json:"id"`
		NamaItem string `json:"Nama_item"`
		HargaItem int `json:"Harga_item"`
		JumlahTerjual int `json:"Jumlah_terjual"`
		Penghasilan int `json:"Penghasilan"`
	}
	
	rows, err := db.Query("SELECT * FROM items")
	if err != nil {
		c.SendString(err.Error())
	}
	defer rows.Close()

	items := []*Item{}
	TotalPenghasilan := 0 
	for rows.Next() {
		var buffer Item
		err := rows.Scan(&buffer.Id,&buffer.NamaItem,&buffer.HargaItem,&buffer.JumlahTerjual)
		if err != nil {
			return c.SendString(err.Error())
		}
		buffer.Penghasilan = buffer.JumlahTerjual * buffer.HargaItem
		TotalPenghasilan += buffer.Penghasilan
		items = append(items, &buffer)
	}

	return c.Render("views/data", fiber.Map{
		"Items": &items,
		"TotalPenghasilan": TotalPenghasilan,
	})
}

func getEdit(c *fiber.Ctx) error {
	type Item struct{
		Id int `json:"id"`
		NamaItem string `json:"Nama_item"`
		HargaItem int `json:"Harga_item"`
	}
	
	rows, err := db.Query("SELECT id, nama_item, harga_item FROM items")
	if err != nil {
		c.SendString(err.Error())
	}
	defer rows.Close()

	items := []*Item{}
	for rows.Next() {
		var buffer Item
		err := rows.Scan(&buffer.Id,&buffer.NamaItem,&buffer.HargaItem)
		if err != nil {
			return c.SendString(err.Error())
		}
		items = append(items, &buffer)
	}

	return c.Render("views/edit", fiber.Map{
		"Items": &items,
	})
}

func getUpdate(c *fiber.Ctx) error {
	type Item struct{
		Id int `json:"id"`
		NamaItem string `json:"Nama_item"`
		HargaItem int `json:"Harga_item"`
	}

	id := c.Params("id")
	
	rows, err := db.Query("SELECT id, nama_item, harga_item FROM items WHERE id=?", id)
	if err != nil {
		c.SendString(err.Error())
	}
	defer rows.Close()


	var item Item
	for rows.Next() {
		err := rows.Scan(&item.Id,&item.NamaItem,&item.HargaItem)
		if err != nil {
			return c.SendString(err.Error())
		}
	}

	return c.Render("views/update", fiber.Map{
		"Id": item.Id,
		"NamaItem": item.NamaItem,
		"HargaItem": item.HargaItem,
	})
}

func Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := db.Exec("DELETE FROM items WHERE id = ?", id)
	if err != nil {
		return c.SendString("Error Deleting Data")
	}

	return c.Redirect("/")
}


func postAddItem(c *fiber.Ctx) error {
	nama_item := c.FormValue("nama_item")
	harga_item := c.FormValue("harga_item")
	harga_item_int,err  := strconv.Atoi(harga_item)
	if err != nil {
		c.SendString("harga harus berupa angka!")
	}
	_, err = db.Exec("INSERT INTO items(nama_item,harga_item,jumlah_terjual) VALUES(?,?,?)", nama_item,harga_item_int,0)
	if err != nil {
		c.SendString(err.Error())
	}
	
	return c.Redirect("/")
}

func postEdit(c *fiber.Ctx) error {
    id := c.FormValue("id")
	nama_item := c.FormValue("nama_item")
	harga_item := c.FormValue("harga_item")
	harga_item_int,err  := strconv.Atoi(harga_item)
	if err != nil {
		c.SendString("harga harus berupa angka!")
	}
	_, err = db.Exec("UPDATE items SET nama_item=?,harga_item=? WHERE id=?", nama_item,harga_item_int,id)
	if err != nil {
		c.SendString(err.Error())
	}
	
	return c.Redirect("/")
}

func postData(c *fiber.Ctx) error {
	type data struct{
		Id interface{} `json:"id"`
		Val interface{} `json:"val"`
	}

	var datas []data

	// fmt.Println(time.Now().Format("02-01-2006")) workin on per day sales :)

	if err := c.BodyParser(&datas); err != nil{
		c.SendString(err.Error())
	}

	for _, v := range datas {
		_, err := db.Exec("UPDATE items SET jumlah_terjual= jumlah_terjual + ? WHERE id=?",v.Val, v.Id)
		if err != nil {
			c.SendString(err.Error())
		}
	}
	
	return c.Redirect("/")
}

