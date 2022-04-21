# GRPC + DAO Support

In this example, we demonstrated how the build-in DAO(Data Access Object, not DAO for web 3.0) support could simplify the repetitive struggling with database CRUD.  

Developers spend most of their lives messing around with mysql/orm and all kinds of ORM packages. In order to perform CRUD, they usually have to do the following:  
- Define a struct/class/whatever model
- Depending on the language, the model may have some fundamental CRUD capabilities (e.g. in Python Django/Flask), or cumbersome heavy classes（e.g. Hiernate/Spring DAO), or nothing at all and you need to use ORM tool to play around (e.g. gorm/xorm).  
- Although we just need CRUD, most ORM tools offer quite a lot flexible methods to query/create, and "smart" solution for Upsert/Soft Delete/etc., which makes the APIs confusing.    
- Developers have to wrap those ORM tools again and again to consolidate to "reusabel" interfaces, if they are really good enough to do that...  

What I like most is the way how Djano does in its model. They have build-in CRUD functionalities. And I'd like to bring similiar features to Skema-Go, as shown in below code:

```
	user := data.Manager().GetDAO(&dao.User{})
	err = user.Upsert(&dao.User{
		UUID: uuid.New().String(),
		Name: req.Msg,
	}, nil, nil)

	if err == nil {
		rs := []dao.User{}
		user.Query(&data.QueryParams{}, &rs)
		result = fmt.Sprintf("total: %d", len(rs))
	} else {
		result = err.Error()
	}
```
Let's go through the code line by line:  
<br/>
<br/>
```
	user := database.Manager().GetDAO(&dao.User{})
```
First, we do NOT design or create a DAO by ourselves, we *Fetch* a DAO object from a database manager. The manager in fact register all DAOs for tables/models you designed, so you can just fetch from the registry, instead create a new one. By this way, we can avoid many repetitive actions when building a DAO object.  

```
err = user.Upsert(&dao.User{
		UUID: uuid.New().String(),
		Name: req.Msg,
	}, nil, nil)
```
After we get the user DAO, we can call Upsert() function to create/update the record. Upsert is a well-known concept that you can either create a new one when there is no conflict (based on unique fields), or update the existing one. In fact, the method could perform three actions: Create brand new one, Update existing one, Update partial fields for existing one. In the above code, we only demonstrated creating new records.  

```
	if err == nil {
		rs := []dao.User{}
		user.Query(&data.QueryParams{}, &rs)
		result = fmt.Sprintf("total: %d", len(rs))
	} else {
		result = err.Error()
	}
```
Query is the most *highlighted* feature for most ORMs. While in our framework, we only provided one solution: map base query key. Our underlying support is provided by gorm, so this is actually the same as described in [gorm query with map condition](https://gorm.io/docs/query.html). We believe this could make most query much easier. If this doesn't meet your needs, you can always use `user.GetDB()` to obtain the underlying gorm.DB instance to perform whatever you want.  

