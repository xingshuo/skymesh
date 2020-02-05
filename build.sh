echo "start building proto..."
protoc --proto_path=./proto/ --go_out=./proto/generate ./proto/*.proto
echo "build proto done."