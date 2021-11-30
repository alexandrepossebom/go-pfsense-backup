# go-pfsense-backup

Simple utility to backup pfsense firewalls.

## Usage

You need file: `go-pfsense-backup.yaml` in `/etc` `/root` or in current directory.

Example of `go-pfsense-backup.yaml`:

```
firewalls:
  - name: Tijolo
    username: alexandre
    password: verysecretpass
    url: https://192.168.100.254
	directory: '/root/backups/'
  - name: Reboco
    username: alexandre
    password: verysecretpass
    url: https://192.168.100.1
	directory: '/root/backups/'
```

Output example:

```
# go run main.go
Using config file: /Users/alexandre/Devel/go-pfsense-backup/go-pfsense-backup.yaml
Starting backup of Tijolo ...done 1.034022458s
Starting backup of Reboco ...done 532.991792ms
Done
```
