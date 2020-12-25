(
    dirs="$@"
    if test -z "$dirs"; then
        dirs="Photos/2020"
    fi

    cd /mnt/photos/gdrive
    for d in $dirs; do
        echo "$(date): pulling $d"
        /mnt/photos/bin/drive pull --quiet --no-prompt --ignore-name-clashes --ignore-conflict $d
        /mnt/photos/bin/fix-dir-time.py $d
    done

    curl 'http://localhost:8081/db/viewer?command=reload'

) > /tmp/cron-pull.log 2>& 1
