tagx=$(date +"%Y%m%d%H%M%S")

dockeruser="wangleilei950325"

dockerimage="chatgpt-telegram-bot"

docker build -t $dockerimage/$dockerimage:$tagx .

if [ $? -eq 0 ]; then
    docker push $dockeruser/$dockerimage:$tagx
    echo "Tag for Docker image is: $tagx"
else
    echo "Docker build failed, image not pushed to Docker Hub"
fi