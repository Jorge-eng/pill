GIT_SHA=$(git rev-parse --short HEAD)
echo "--> Building..."
go build -o bin/pill-osx -ldflags="-X main.FactoryKey=${PILL_KEY} -X main.GitSha=${GIT_SHA}"
echo "--> Copying pill to /usr/local/bin/"
cp bin/pill-osx /usr/local/bin/pill
echo "--> Done"