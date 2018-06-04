data_dir=/data/factom/private

./factomd -config $data_dir/factomd.conf -factomhome $data_dir -network LOCAL -loglvl info "$@"
