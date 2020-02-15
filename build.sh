echo "start building proto..."
protoc --proto_path=./proto/ --go_out=./proto/generate ./proto/*.proto
echo "build proto done."

cd nameserver/bootstrap
rm nameserver.exe
go build -gcflags "-N -l" -mod=vendor -o nameserver.exe main.go
cd ../../
cp nameserver/bootstrap/nameserver.exe examples/helloworld/nameserver/main.exe
cp nameserver/bootstrap/nameserver.exe examples/nameservice/nameserver/main.exe
echo "build nameserver done."

cd examples/helloworld/greeter_client
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build greeter client done."

cd examples/helloworld/greeter_server
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build greeter server done."

cd examples/nameservice/client
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build nameservice client done"

cd examples/nameservice/server
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build nameservice server done"