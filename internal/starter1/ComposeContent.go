package starter1

import "fmt"

func ComposeContent(project, appDir string) string {
	service := fmt.Sprintf("%v-%v", appDir, project)
	return fmt.Sprintf(`name: %v-${APP_ENV}
services:
  %v:
    build:
      target: ${APP_ENV}
      context: ../%v
      dockerfile: ../docker/Dockerfile
      args:
        - NODE_VERSION=${NODE_VERSION}
    container_name: %v_${APP_ENV}_app
    image: %v-app:${APP_ENV}
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=%v-nseven"
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
      - %v-nseven
      - %v

networks:
  %v-nseven:
    external: true
  %v:
    driver: bridge
`, appDir, project, project, appDir, appDir,
		appDir, service, service, service, service, service, service,
		project, project, appDir, appDir, appDir, appDir)
}

func ComposePreprodContent(project, appDir string) string {
	service := fmt.Sprintf("%v-%v", appDir, project)
	return fmt.Sprintf(`name: %v-${APP_ENV}
services:
  %v:
    build:
      target: ${APP_ENV}
      context: ../%v
      dockerfile: ../docker/Dockerfile
      args:
        - NODE_VERSION=${NODE_VERSION}
    container_name: %v_${APP_ENV}_app
    image: %v-app:${APP_ENV}
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=%v-nseven"
      - "traefik.http.routers.%v.rule=${HOST_TRAEFIK_APP}"
      - "traefik.http.routers.%v.entrypoints=websecure"
      - "traefik.http.routers.%v.tls=true"
      - "traefik.http.routers.%v.tls.certresolver=default"
      - "traefik.http.services.%v.loadbalancer.server.port=${PORT}"
    volumes:
      - ../%v:/app
    env_file:
      - ../%v/.env
    networks:
      - %v-nseven
      - %v

networks:
  %v-nseven:
    external: true
  %v:
    driver: bridge
`, appDir, project, project, appDir, appDir,
		appDir, service, service, service, service, service,
		project, project, appDir, appDir, appDir, appDir)
}

func ComposeProdContent(project, appDir string) string {
	service := fmt.Sprintf("%v-%v", appDir, project)
	return fmt.Sprintf(`name: %v-${APP_ENV}
services:
  %v:
    build:
      target: ${APP_ENV}
      context: ../%v
      dockerfile: ../docker/Dockerfile
      args:
        - NODE_VERSION=${NODE_VERSION}
    container_name: %v_${APP_ENV}_app
    image: %v-app:${APP_ENV}
    labels:
      - "traefik.enable=true"
      - "traefik.docker.network=%v-nseven"
      - "traefik.http.routers.%v.rule=${HOST_TRAEFIK_APP}"
      - "traefik.http.routers.%v.entrypoints=websecure"
      - "traefik.http.routers.%v.tls=true"
      - "traefik.http.routers.%v.tls.certresolver=default"
      - "traefik.http.services.%v.loadbalancer.server.port=${PORT}"
    volumes:
      - ../%v:/app
    env_file:
      - ../%v/.env
    networks:
      - %v-nseven
      - %v

networks:
  %v-nseven:
    external: true
  %v:
    driver: bridge
`, appDir, project, project, appDir, appDir,
		appDir, service, service, service, service, service,
		project, project, appDir, appDir, appDir, appDir)
}
