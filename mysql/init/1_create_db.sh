#!/bin/bash -eu

mysql=( mysql --protocol=socket -uroot -p"${MYSQL_ROOT_PASSWORD}" )

"${mysql[@]}" <<-EOSQL
    CREATE DATABASE IF NOT EXISTS micro_user default character set utf8;
    GRANT ALL ON micro_user.* TO '${MYSQL_USER}'@'%' ;
EOSQL