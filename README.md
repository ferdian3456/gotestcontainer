## Go TestContainer

A simple implementation of using go test container. I personally believe that using integration test is better than using mock and also its already enough to test the backend api rather than create a integration test like this with unit test.

## Case 1: Basic

Run basic test with testcontainer

```bash
go test -v
```

## Case 2: Singleton

This case means that all the test were using the same database, so that it can reduce the time to run the test.

```bash
go test -v ./singleton
```

How it works:

1. When you run the test it actuall call the "TestMain(m \*testing.M)"" first and then run the actual test(xxx_test.go)
2. This function will call "run" function that setup the database connection and url. After that it will set the value of global variable "dbPool" with the database connection that you get from the "run" function. The reason behind putting all the setup in "run" function is because it needs defer the database connection pool and the container after the test is done. You cant do "os.Exit" while also use defer within the same function.

## Case 3: Parallel

This case means that all the test were running in parallel with different database(postgres database) within the same postgres container, so that it can reduce the time to run the test.

```bash
go test -v ./parallel
```

How it works:

1. When you run the test it actuall call the "TestMain" function first and then run the actual test(xxx_test.go)
2. This function will call "run" function that setup the database connection and url.
3. All the test were running in parallel, and for every test it will create a new database in postgres by calling "SetupTestDatabase" that is dedicated for that test. So if there are 10 test, it will create 10 database in postgres. After the test function is finished it will drop the database, dont forget that there is "t.Cleanup" in "SetupTestDatabase", t.Cleanup will run after the test function is finished and you know that every test were calling that "SetupTestDatabase" function.

## Notes

The reason why in singleton case there is no need to define the conn string as global variable is because the test were just querying to database using "dbpool" that was created in "TestMain", in third case it need conn string because it was required to create a new db pool to create a new database and drop the database.

## FAQ

1. What happened if i have a case where i need to perform some action before running the test, for example like to create a user account, and join server first before start to test the create post api?

**Answer:**
You should perform those actions within the test function itself or through a setup helper (often called a "Factory" or "Seeder"). This ensures that each test is **self-contained** and doesn't depend on other tests.

Because we use **Case 3 (Parallel Isolation)**, each test has its own unique database. Even if multiple tests create the same user (e.g., "admin"), they won't clash because they are in separate database instances.

So for example if you need to test CreateServerPost that means before you hit the api create server post u need to create a user or register the user first to get the jwt or something that was required for you to hit the protected api, after that you need to join the server and then finally test the real case.
