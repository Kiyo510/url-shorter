.PHONY: swagger
swagger:
	docker run -p 8081:8080 -e SWAGGER_JSON=/docs/swagger/swagger.yml -v $(PWD)/docs:/docs swaggerapi/swagger-ui
