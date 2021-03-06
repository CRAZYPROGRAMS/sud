package core

import (
	"database/sql"
	"errors"
)

func (core *Core) addColumn(tx *transaction, info *PoleTableInfo) error {
	var err error
	PoleDBType := ""
	var Q2 string = ""
	switch info.PoleInfo.GetPoleType() {
	case "StringValue":
		PoleDBType = `text NULL`
		break
	case "Int64Value":
		PoleDBType = `bigint NULL`
		break
	case "DateValue":
		PoleDBType = `date NULL`
		break
	case "DateTimeValue":
		PoleDBType = `timestamp NULL`
		break
	case "RecordLinkValue":
		PoleDBType = `uuid NULL`
		Q2 = `ALTER TABLE "` + info.TableName + `" ADD CONSTRAINT "` + info.TableName + `_fk_` + info.PoleName + `" FOREIGN KEY (t) REFERENCES public."Record" ("__RecordUID") MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION;`
		break
	default:
		return errors.New("pole type error")
	}
	if _, err = tx.Exec(`ALTER TABLE "` + info.TableName + `" ADD "` + info.PoleName + `" ` + PoleDBType); err != nil {
		return err
	}
	if Q2 != "" {
		if _, err = tx.Exec(Q2); err != nil {
			return err
		}
	}
	if info.PoleInfo.GetIndexType() == "Unique" {
		if _, err = tx.Exec(`CREATE UNIQUE INDEX "UIndex_` + info.TableName + `" ON "` + info.TableName + `" ("` + info.PoleName + `" ASC NULLS LAST)`); err != nil {
			return err
		}
	}
	if info.PoleInfo.GetIndexType() == "Index" {
		if _, err = tx.Exec(`CREATE INDEX "UIndex_` + info.TableName + `" ON "` + info.TableName + `" ("` + info.PoleName + `" ASC NULLS LAST)`); err != nil {
			return err
		}
	}
	return nil

}
func (core *Core) CheckConfiguration(TransactionUID string, ConfigurationName string) error {
	var err error
	var ok bool
	var tx *transaction
	var config *Configuration
	if tx, err = core.getTransaction(TransactionUID); err != nil {
		return err
	}
	if config, err = core.LoadConfiguration(ConfigurationName); err != nil {
		return err
	}
	TablePoles := map[string]map[string]*PoleTableInfo{}
	for _, poles := range config.polesInfo {
		for _, pi := range poles {
			pti := &PoleTableInfo{}
			if err = pti.FromPoleInfo(pi); err != nil {
				return err
			}
			var ti map[string]*PoleTableInfo
			ti, ok = TablePoles[pti.TableName]
			if !ok {
				ti = map[string]*PoleTableInfo{}
				TablePoles[pti.TableName] = ti
			}
			ti[pti.PoleName] = pti
		}
	}
	Tables := map[string]map[string]string{}
	var rows1, rows2 *sql.Rows

	if rows1, err = tx.Query(`SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_SCHEMA='public'`); err != nil {
		return err
	}
	defer rows1.Close()
	for rows1.Next() {
		var TABLE_NAME string
		if err = rows1.Scan(&TABLE_NAME); err != nil {
			return err
		}
		Tables[TABLE_NAME] = map[string]string{}
	}
	if rows2, err = tx.Query("SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='public'"); err != nil {
		return err
	}
	defer rows2.Close()
	for rows2.Next() {
		var TABLE_NAME, COLUMN_NAME, DATA_TYPE string
		if err = rows2.Scan(&TABLE_NAME, &COLUMN_NAME, &DATA_TYPE); err != nil {
			return err
		}
		Tables[TABLE_NAME][COLUMN_NAME] = DATA_TYPE
	}
	for TableName, ti := range TablePoles {
		var poles map[string]string
		if poles, ok = Tables[TableName]; !ok {
			if _, err = tx.Exec(`CREATE TABLE "` + TableName + `" ( "__RecordUID" uuid NOT NULL, CONSTRAINT "` + TableName + `_pk_record" PRIMARY KEY ("__RecordUID"),   CONSTRAINT "` + TableName + `_fk_record" FOREIGN KEY ("__RecordUID") REFERENCES "Record" ("__RecordUID") MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION)`); err != nil {
				return err
			}
			poles = map[string]string{}
		}
		for ColumnName, pi := range ti {
			if _, ok = poles[ColumnName]; !ok {
				if err = core.addColumn(tx, pi); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (core *Core) CreateDatabase() error {
	var err error
	tid, err := core.BeginTransaction()
	if err != nil {
		return err
	}
	defer core.RollbackTransaction(tid)
	tx, _ := core.getTransaction(tid)

	//dbname := con.getDBName()
	if _, err = tx.Exec(`
CREATE TABLE "Record"
(
  "__RecordUID" uuid NOT NULL,
  "RecordType" text,
  "Readonly" boolean NOT NULL,
  "DeleteRecord" boolean NOT NULL,
  CONSTRAINT "pk_record" PRIMARY KEY ("__RecordUID")
)	
	`); err != nil {
		return err
	}
	core.CommitTransaction(tid)
	return nil
}
