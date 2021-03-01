# grpcFlags

Simple grpc framework for local based ctf's

Structure of framework:
1. `server` - main server for stored users and flags
2. `client` - client for auth, and submit flags
3. `adminClient` - client for admin access to dashboard
4. `services` - example of service that pull different flag for each user

Features:
* flag's encrypted transmission 
* unique flag for each user

Deploy:
1. Build server

You can run it on cloud
```
go run server/main.go
```
2. Auth users

For a register local tasks, you need tou auth
```
go run client/main.go
```

3. Run example task

Sql injection task will be on localhost:8000

```
go run services/SQLInjection/server.go
```
service will be registered with credentials of user, and admin could abel to see progress in admin app

4. Check Users

You can see users stats in adminClient
```
go run adminClient/main.go
```