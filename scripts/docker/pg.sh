docker run -d \
  --name my-postgres \
  -e POSTGRES_USER=admin \
  -e POSTGRES_PASSWORD=123456 \
  -p 5432:5432 \
  -v ~/data/dws/pg:/var/lib/postgresql \
  postgres:18.0