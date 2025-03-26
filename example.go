package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
)

// 定义一个全局对象db
var db *sql.DB
var u user
var users []user

type user struct {
	Id   int
	Age  int
	Name string
}

// 定义一个初始化数据库的函数
func initDB() (err error) {
	// DSN:Data Source Name
	dsn := "root:root@tcp(127.0.0.1:3306)/059404?charset=utf8mb4&parseTime=True"
	// 不会校验账号密码是否正确
	// 注意！！！这里不要使用:=，我们是给全局变量赋值，然后在main函数中使用全局变量db
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// 尝试与数据库建立连接（校验dsn是否正确）
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

// 使用 CORS 中间件

func main() {
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 允许所有来源的请求
	config.AllowMethods = []string{"*"} // 设置允许的方法
	config.AllowHeaders = []string{"*"} // 设置允许的头部参数
	config.AllowCredentials = true      // 是否允许携带cookie等身份认证信息

	router.Use(cors.New(config))

	err := initDB() // 调用输出化数据库的函数

	if err != nil {
		fmt.Printf("init db failed,err:%v\n", err)
		return
	}

	router.GET("/queryOne", func(c *gin.Context) {
		queryRow()

		c.JSON(http.StatusOK, u)
	})
	router.GET("/queryAll", func(c *gin.Context) {
		queryAll()

		c.JSON(http.StatusOK, users)
	})

	router.GET("/query", func(c *gin.Context) {
		queryRowDemo()

		c.JSON(http.StatusOK, u)
	})
	router.POST("/insert", func(c *gin.Context) {

		bodyByts, err := ioutil.ReadAll(c.Request.Body)

		if err != nil {
			// 返回错误信息
			c.String(http.StatusBadRequest, err.Error())
			// 执行退出
			c.Abort()
		}

		jsonStr := string(bodyByts)

		var result map[string]interface{}

		err1 := json.Unmarshal([]byte(jsonStr), &result)
		if err1 != nil {
			log.Fatalf("JSON Unmarshalling failed: %s", err1)
		}
		fmt.Printf("jsonstr", jsonStr)
		fmt.Println("666", result)
		// 定义一个结构体，字段名与JSON中的键名对应
		type Person struct {
			Name string `json:"name"`
			Age  string `json:"age"`
		}
		var person Person
		err2 := json.Unmarshal([]byte(jsonStr), &person)
		if err2 != nil {
			log.Fatalf("JSON Unmarshalling failed: %s", err2)
		}

		// 输出结果
		fmt.Printf("Name: %s, Age: %d\n", person.Name, person.Age)

		name := person.Name
		age, err3 := strconv.Atoi(person.Age)
		if err3 != nil {
			log.Fatalf("JSON Unmarshalling failed: %s", err2)
		}

		insertRowDemo(name, age)

		c.JSON(http.StatusOK, u)
	})
	router.POST("/update", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Query("id"))
		if err != nil {
			fmt.Println("转换错误:", err)
		} else {
			fmt.Println("转换后的整型:", id)

		}
		updateRowDemo(id)

		c.JSON(http.StatusOK, u)
	})
	router.POST("/delete", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Query("id"))
		if err != nil {
			fmt.Println("转换错误:", err)
		} else {
			fmt.Println("转换后的整型:", id)

		}
		deleteRowDemo(id)
		c.JSON(http.StatusOK, u)
	})

	// 监听并在 0.0.0.0:8080 上启动服务
	router.Run(":8080")
}

// // 查询单条数据示例
func queryRowDemo() {
	sqlStr := "select id, name, age from user where id=?"

	// 非常重要：确保QueryRow之后调用Scan方法，否则持有的数据库链接不会被释放
	err := db.QueryRow(sqlStr, 1).Scan(&u.Id, &u.Name, &u.Age)
	if err != nil {
		fmt.Println("scan failed, err:%v\n", err)
		return
	}
	fmt.Println("id:%d name:%s age:%d\n", u.Id, u.Name, u.Age)

}

func queryRow() {
	sqlStr := "select id, name, age from user where id>0 limit 1"

	// 非常重要：确保QueryRow之后调用Scan方法，否则持有的数据库链接不会被释放
	err := db.QueryRow(sqlStr).Scan(&u.Id, &u.Name, &u.Age)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return
	}

}

func queryAll() {
	sqlStr := "select id, name, age from user where id>1 "

	// 非常重要：确保QueryRow之后调用Scan方法，否则持有的数据库链接不会被释放
	rows, err := db.Query(sqlStr)
	if err != nil {
		return
	}
	defer rows.Close()
	users = nil
	for rows.Next() {
		var item user
		if err := rows.Scan(&item.Id, &item.Name, &item.Age); err != nil {
			return
		}

		users = append(users, item)
	}

}

// 插入数据
func insertRowDemo(name string, age int) {
	sqlStr := "insert into user(name, age) values (?,?)"
	ret, err := db.Exec(sqlStr, name, age)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", sqlStr)
		fmt.Printf("insert failed, err:%v\n", name)
		fmt.Printf("insert failed, err:%v\n", age)
		return
	}
	theID, err := ret.LastInsertId() // 新插入数据的id
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
	fmt.Printf("insert success, the id is %d.\n", theID)
}

// 更新数据
func updateRowDemo(i int) {
	sqlStr := "update user set age=? where id = ?"
	ret, err := db.Exec(sqlStr, 39, i)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return
	}
	n, err := ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("update success, affected rows:%d\n", n)
}

// 删除数据
func deleteRowDemo(id int) {
	sqlStr := "delete from user where id = ?"
	ret, err := db.Exec(sqlStr, id)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return
	}
	n, err := ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
	fmt.Printf("delete success, affected rows:%d\n", n)
}
