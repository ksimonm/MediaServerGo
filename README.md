# Mediaserver-go

Image and API Platform(PoC)

## Setup local development

  ```bash
  mkdir config && cp config.example.json config/config.json 
  ```

  ```bash
  go run server.go
  ```

  ```bash
  aws --endpoint-url http://localhost:8080 s3 cp image.jpg s3://mybucket
  ```

  Open in browser
  ```
  http://localhost:8080/mybucket/image.jpg?h=200&w=200&q=1
  ```