package stage1

import "fmt"

func ComposeContent(nameApp, nameFolderProject string) string {
	service := fmt.Sprintf("%v-%v", nameFolderProject, nameApp)
	return fmt.Sprintf(`name: %v-${APP_ENV}
services:
  app:
     build:
       target: dev
       context: ../%v
       dockerfile: ../docker/%v.dockerfile
       args:
         - NODE_VERSION=${NODE_VERSION}
     container_name: %v_${APP_ENV}_%v
     image: %v-%v:${APP_ENV}
     labels:
       - "traefik.enable=true"
       - "traefik.docker.network=traefik-nseven"
       - "traefik.http.routers.%v.rule=${HOST_TRAEFIK_APP}"
       - "traefik.http.routers.%v.entrypoints=websecure"
       - "traefik.http.routers.%v.tls=true"
       - "traefik.http.routers.%v.tls.certresolver=default"
       - "traefik.http.services.%v.loadbalancer.server.port=${PORT}"
       - "traefik.http.services.%v.loadbalancer.server.scheme=http"
     volumes:
       - ../%v:/app
     env_file:
       - ../%v/.env
     networks:
       - traefik-nseven
       - %v

networks:
  traefik-nseven:
     external: true
  %v:
     driver: bridge
`, nameFolderProject, nameApp, nameApp, nameFolderProject, nameApp, nameFolderProject, nameApp,
		service, service, service, service, service, service,
		nameApp, nameApp, nameFolderProject, nameFolderProject)
}

func ComposePreprodContent(nameApp, nameFolderProject string) string {
	service := fmt.Sprintf("%v-%v", nameFolderProject, nameApp)
	return fmt.Sprintf(`name: %v-${APP_ENV}
services:
  app:
    image: ghcr.io/nsevendev/%v/app:${IMAGE_TAG}
    container_name: %v_${APP_ENV}_app
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=traefik-nseven"
      - "traefik.http.routers.%v-${APP_ENV}.rule=${HOST_TRAEFIK_APP}"
      - "traefik.http.routers.%v-${APP_ENV}.entrypoints=websecure"
      - "traefik.http.routers.%v-${APP_ENV}.tls=true"
      - "traefik.http.routers.%v-${APP_ENV}.tls.certresolver=le"
      - "traefik.http.services.%v-${APP_ENV}.loadbalancer.server.port=${PORT}"
    environment:
      NODE_ENV: production
    env_file:
      - ../%v/.env
    networks:
      - traefik-nseven
      - %v

networks:
  traefik-nseven:
    external: true
  %v:
    driver: bridge
`, nameFolderProject, nameFolderProject, nameFolderProject,
		service, service, service, service, service,
		nameApp, nameFolderProject, nameFolderProject)
}

func ComposeProdContent(nameApp, nameFolderProject string) string {
	service := fmt.Sprintf("%v-%v", nameFolderProject, nameApp)
	return fmt.Sprintf(`name: %v-${APP_ENV}
services:
  app:
    image: ghcr.io/nsevendev/%v/app:${IMAGE_TAG}
    container_name: %v_${APP_ENV}_app
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=traefik-nseven"
      - "traefik.http.routers.%v-${APP_ENV}.rule=${HOST_TRAEFIK_APP}"
      - "traefik.http.routers.%v-${APP_ENV}.entrypoints=websecure"
      - "traefik.http.routers.%v-${APP_ENV}.tls=true"
      - "traefik.http.routers.%v-${APP_ENV}.tls.certresolver=le"
      - "traefik.http.services.%v-${APP_ENV}.loadbalancer.server.port=${PORT}"
    environment:
      NODE_ENV: production
    env_file:
      - ../%v/.env
    networks:
      - traefik-nseven
      - %v

networks:
  traefik-nseven:
    external: true
  %v:
    driver: bridge
`, nameFolderProject, nameFolderProject, nameFolderProject,
		service, service, service, service, service,
		nameApp, nameFolderProject, nameFolderProject)
}
