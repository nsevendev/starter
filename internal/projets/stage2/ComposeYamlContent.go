package stage2

import "fmt"

func ComposeYamlContent(nameFolderProject string) string {
	return fmt.Sprintf(`name: %s-${APP_ENV}
services:
  front:
    build:
      target: ${APP_ENV}
      context: ../front
      dockerfile: ../docker/front.dockerfile
    container_name: %s_${APP_ENV}_front
    image: %s-front:${APP_ENV}
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=traefik-nseven"
      - "traefik.http.routers.%s-front.rule=${HOST_TRAEFIK_FRONT}"
      - "traefik.http.routers.%s-front.entrypoints=websecure"
      - "traefik.http.routers.%s-front.tls=true"
      - "traefik.http.routers.%s-front.tls.certresolver=default"
      - "traefik.http.services.%s-front.loadbalancer.server.port=${PORT}"
      - "traefik.http.services.%s-front.loadbalancer.server.scheme=http"
    volumes:
      - ../front:/app
    env_file:
      - ../front/.env
    networks:
      - traefik-nseven
      - %s

  api:
    build:
      target: ${APP_ENV}
      context: ../api
      args:
        SERVICE: api
      dockerfile: ../docker/api.dockerfile
    container_name: %s_${APP_ENV}_api
    image: %s-api:${APP_ENV}
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=traefik-nseven"
      - "traefik.http.routers.%s-api.rule=${HOST_TRAEFIK_API}"
      - "traefik.http.routers.%s-api.entrypoints=websecure"
      - "traefik.http.routers.%s-api.tls=true"
      - "traefik.http.routers.%s-api.tls.certresolver=default"
      - "traefik.http.services.%s-api.loadbalancer.server.port=${PORT}"
      - "traefik.http.services.%s-api.loadbalancer.server.scheme=http"
    volumes:
      - ../api:/app
    env_file:
      - ../api/.env
    networks:
      - traefik-nseven
      - %s
    depends_on:
      - db

  db:
    image: mongo:7
    container_name: %s_${APP_ENV}_db
    restart: unless-stopped
    volumes:
      - %s_dev_db:/data/db
      - ../docker/mongo-init:/docker-entrypoint-initdb.d
    ports:
      - "${DB_PORT_EX:-27017}:27017"
    networks:
      - traefik-nseven
      - %s

networks:
  %s:
    driver: bridge
  traefik-nseven:
    external: true

volumes:
  %s_dev_db:
`,
		nameFolderProject, // name project
		nameFolderProject, // container_name front
		nameFolderProject, // image front
		nameFolderProject, // router front rule
		nameFolderProject, // router front entrypoints
		nameFolderProject, // router front tls
		nameFolderProject, // router front tls.certresolver
		nameFolderProject, // service front loadbalancer
		nameFolderProject, // service front loadbalancer scheme
		nameFolderProject, // network front

		nameFolderProject, // container_name api
		nameFolderProject, // image api
		nameFolderProject, // router api rule
		nameFolderProject, // router api entrypoints
		nameFolderProject, // router api tls
		nameFolderProject, // router api tls.certresolver
		nameFolderProject, // service api loadbalancer
		nameFolderProject, // service api loadbalancer scheme
		nameFolderProject, // network api

		nameFolderProject, // container_name db
		nameFolderProject, // volume db
		nameFolderProject, // network db

		nameFolderProject, // network name
		nameFolderProject, // volume db
	)
}
