# speech-and-text
google cloudの `speech to text`と `text to speech` に関するリポジトリ

## terraform
```
terraform init
```

```
docker build -t gcr.io/YOUR_PROJECT_ID/speech-and-text .
docker push gcr.io/YOUR_PROJECT_ID/speech-and-text
```

```
terraform apply
```