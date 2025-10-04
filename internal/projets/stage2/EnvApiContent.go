package stage2

import "fmt"

func EnvApiContent(nameFolderProject string, hostApi string) string {
	return fmt.Sprintf(`APP_ENV=dev
# connexion db en mode dev
DB_NAME=%v_dev
DB_URI=mongodb://db:27017
# pour les logs
HOST_TRAEFIK_API=Host(`+"`%v`"+`)
# pour start server
PORT=3000
# info pour port db
DB_PORT_EX=27018
# clef secrete jwt
JWT_SECRET_KEY=supersecretkey

# S3 cloudflare R2
R2_ACCOUNT_ID=
R2_ACCESS_KEY_ID=
R2_SECRET_ACCESS_KEY=
R2_BUCKET_NAME=

# pour redis
REDIS_ADDR=redis:6379

# mailer
MAIL_HOST=sandbox.smtp.mailtrap.io
MAIL_PORT=587
MAIL_USER=79a16633028c7e
MAIL_PASS=a39213807af5f5
MAIL_FROM=john@example.com

# host cors
CORS_DEV_APP=https://%v
CORS_PREPROD_APP=https://nseven.woopear.fr
CORS_PROD_APP=https://nseven.woopear.fr
`, nameFolderProject, hostApi, hostApi)
}
