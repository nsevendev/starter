package stage2

import "fmt"

func ComposeProdYamlContent(nameFolderProject string) string {
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
      - "traefik.http.routers.%s-front.tls.certresolver=le"
      - "traefik.http.services.%s-front.loadbalancer.server.port=${PORT}"
    env_file:
      - ../front/.env
    networks:
      - traefik-nseven
      - %s
    depends_on:
      - api
    restart: unless-stopped

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
      - "traefik.http.routers.%s-api.tls.certresolver=le"
      - "traefik.http.services.%s-api.loadbalancer.server.port=${PORT}"
    env_file:
      - ../api/.env
    volumes:
      - ../api/cli:/app/cli
      - ../api/go.mod:/app/go.mod
      - ../api/go.sum:/app/go.sum
      - ../api/internal:/app/internal
    networks:
      - traefik-nseven
      - %s
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: mongo:7
    container_name: %s_${APP_ENV}_db
    restart: unless-stopped
    volumes:
      - %s_${APP_ENV}_db:/data/db
      - ../docker/mongo-init:/docker-entrypoint-initdb.d
    networks:
      - traefik-nseven
      - %s
    environment:
      - MONGO_INITDB_DATABASE=${DB_NAME}

networks:
  traefik-nseven:
    external: true
  %s:
    driver: bridge

volumes:
  %s_prod_db:
`,
		nameFolderProject, // name
		nameFolderProject, // container_name front
		nameFolderProject, // image front
		nameFolderProject, // router front rule
		nameFolderProject, // router front entrypoints
		nameFolderProject, // router front tls
		nameFolderProject, // router front tls.certresolver
		nameFolderProject, // service front loadbalancer
		nameFolderProject, // network front

		nameFolderProject, // container_name api
		nameFolderProject, // image api
		nameFolderProject, // router api rule
		nameFolderProject, // router api entrypoints
		nameFolderProject, // router api tls
		nameFolderProject, // router api tls.certresolver
		nameFolderProject, // service api loadbalancer
		nameFolderProject, // network api

		nameFolderProject, // container_name db
		nameFolderProject, // volume db
		nameFolderProject, // network db

		nameFolderProject, // network name
		nameFolderProject, // volume name
	)
}
