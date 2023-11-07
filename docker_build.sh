VERSION=`git describe --tags`
BUILD_DATE=`date -u +%d%m%y.%H%M%S`
docker build --build-arg "version=$VERSION build_date=$build_date" --tag=svsignalnew --file Dockerfile .