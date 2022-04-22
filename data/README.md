# Config Driven Database Setup

## The Pain of Data Setup in Backend Development
In most traditional development processes, you'll have to do quite some repetitive works to get your database setup. For example, the below code is pretty common:  
```
user := os.Getenv("user")
pass := os.Getenv("pass")
host := os.Getenv("host")
port := os.Getenv("port")
dbname := os.Getenv("dbname")
charset := os.Getenv("charset")
dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", user, pass, host, port, dbname, charset)
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

And this is barely the minimum code to setup a mysql connection. In a produciton environment, we need to do much more like verifying if the values are valid, and check if all tables have been set (migrated).  

Furthermore, we need to craft something like a DAO object(struct) for CRUD operations (Yes, gorm/xorm and other libs provides fantastic APIs to do all kinds of database operations. They even support at least 4~5 different ways to do a simple query...but all we need is just straightforward CRUD. Especially in this microservice and distributed data service world, we don't really care about transactions/joint tables/foreign keys/etc. If the data layer gets really complicated, we'd better off building a dedicated data service, instead of relying on local orm libraries.)  

*How does skema-go can help?*

## Config driven, NOT code driven!  

Skema-Go provides the data APIs with two rules bearing in mind:  
- Config driven! User (Developer) shouldn't see any explicitly setup code when developing their service  
- Seperate DAO and data model implmementing. Developer defines data model, and that's it! Skema-Go provides CRUDs and all underlying optimization (e.g.     cache support, distributed lock --- coming soon)  
### Data config
Let's see how skema-go uses config file to define a database (In production, this could be saved as text in any remote data or simpley environment variable):  
```
# config/database.yaml
database:
  db1:
     type: mysql       # type of database
     username: root    # account information
     password: 123456
     host: localhost   # databse host 
     port: 3306        # port
     dbname: test      # database name to connect to
     automigrate: true # whether we should migrate automatically or not
     models:           # models(tables) to be initiated (better used with atomigrate flag)
         - User:       # model name
         - Address:
             package: grpc-dao/internal/model   # package name in case your models are in diferent packages
  db2:
     type: sqlite
     filepath: default.db
     dbname: test

```
The above config defines all the properties for a mysql database. You can actually define multiple database connections as above by adding `db2` and the properties. 

### Initialize data in your main() function
How about the code? Let's check out:  
```
func main() {
	data.InitWithFile("./config/database.yaml", "database")

	grpcSrv := grpcmux.NewServer(
		grpc.ChainUnaryInterceptor(Interceptor1(), Interceptor2()),
	)
...
}
```
We added only One Line to initailize everything!  
Everything? You may wonder. So after this one line setup, how do I do CRUDs?  

### Built in DAO (Data Access Object) for CRUD
Let's say you defined a User struct as below:  
```
import (
	"github.com/skema-dev/skema-go/data"
	"gorm.io/gorm"
)

func init() {
	data.RegisterModelType(&User{})  // since golang doesn't have reflect.Typeof(string) method, 
                                     // we have to add this, so the config driven appraoch could work
}

type User struct {
	gorm.Model

	UUID   string `gorm:"column:uuid;index;unique;"`
	Name   string
	Nation string
	City   string
}

func (User) TableName() string {
	return "user"
}
```

Now when you want to do CRUD for the user table, simply fetch it as below:  
```
	user := data.Manager().GetDAO(&model.User{})
```

Just send in the model interface you need, and you'll get a DAO with CRUD capabilities! Now you can use it immediately:  
```
	err = user.Upsert(&model.User{
		UUID: uuid.New().String(),
		Name: req.Msg,
	}, nil, nil)

	if err == nil {
		rs := []model.User{}
		user.Query(&data.QueryParams{}, &rs)
		result = fmt.Sprintf("total: %d", len(rs))

		if len(rs) > 3 {
			err = user.Delete("name like 'user%'")
			if err != nil {
				logging.Errorf(err.Error())
			}
		}

	} 
```

See, putting all your tedious database configuration in a yaml file, and simly add an init() function in your model definition. Load the config, and you are all set!!

Checkout the `grpc-dao` sample and the unit tests code in `/data/manager_test.go` for more details.
