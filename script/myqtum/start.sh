data_dir=/data/qtum/private

./qtumd --datadir=$data_dir --txindex --reindex --logevents -record-log-opcodes --server --regtest --printtoconsole
