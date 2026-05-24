# Backup and Restore Ops Guide

This guide is the practical checklist for staging and pilot production. It complements the full backup runbook in `docs/backup_restore_runbook_umkm_commerce_os_hybrid.md`.

## Backup schedule

Minimum pilot schedule:

- Daily PostgreSQL backup at 02:00 server time.
- Manual backup before every deploy.
- Manual backup before risky migration or data repair.
- Copy backups outside the VPS at least daily.

Example cron:

```cron
0 2 * * * cd /home/deploy/umkm-commerce-os && ./deploy/scripts/backup-db.sh >> ./backups/backup.log 2>&1
```

Retention is controlled by:

```bash
BACKUP_RETENTION_DAYS=14 ./deploy/scripts/backup-db.sh
```

Use a longer retention window for the first pilot if disk space allows.

## Where backups are stored

Default local path:

```txt
backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
```

Local backups are not enough for real recovery. Sync them to another VPS, S3/R2, or another controlled storage location. Do not place backups under public web roots.

## Run a backup

```bash
./deploy/scripts/backup-db.sh
```

The script:

- creates a timestamped `pg_dump`
- compresses the output
- verifies gzip integrity
- removes old backups based on retention
- avoids printing database passwords

## Verify a backup file

```bash
ls -lh backups/database/daily
gzip -t backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
```

The file must be non-empty and gzip validation must pass.

## Restore test procedure

Never test restore directly on production first.

1. Prepare a disposable staging database.
2. Copy the backup file to staging.
3. Run:

   ```bash
   ./deploy/scripts/restore-db.sh --yes backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
   ```

4. Validate table counts:

   ```sql
   SELECT count(*) FROM tenants;
   SELECT count(*) FROM stores;
   SELECT count(*) FROM products;
   SELECT count(*) FROM orders;
   SELECT count(*) FROM stock_movements;
   ```

5. Run smoke checks:
   - login
   - open dashboard
   - open public store
   - inspect one order
   - inspect stock snapshot and movement

## Production restore warning

Production restore is destructive. Do not run it blindly.

Before restoring production:

1. Confirm incident severity and approval.
2. Stop writes if possible.
3. Backup the current damaged state for investigation.
4. Restore to staging first if time allows.
5. Use an explicit backup file.

Interactive restore:

```bash
./deploy/scripts/restore-db.sh backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
```

Non-interactive restore, for controlled emergency scripts only:

```bash
./deploy/scripts/restore-db.sh --yes backups/database/daily/umkm_os_db_YYYYMMDD_HHMMSS.sql.gz
```

