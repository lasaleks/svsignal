VERSION=`git describe --tags`
docker build --tag=svsignalnew:$VERSION .