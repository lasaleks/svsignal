package main

import (
	"database/sql"
	"fmt"
)

var sql_create_svsignal_fvalue = []string{
	`CREATE TABLE svsignal_fvalue (
	id int(11) NOT NULL,
	signal_id int(11) NOT NULL,
	utime int(11) NOT NULL,
	value double NOT NULL,
	offline tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
	`ALTER TABLE svsignal_fvalue
	ADD PRIMARY KEY (id),
	ADD KEY signal_id (signal_id);`,
	`ALTER TABLE svsignal_fvalue
MODIFY id int(11) NOT NULL AUTO_INCREMENT;`,
	`ALTER TABLE svsignal_fvalue
ADD CONSTRAINT svsignal_fvalue_ibfk_1 FOREIGN KEY (signal_id) REFERENCES svsignal_signal (id);`,
}

var sql_create_svsignal_ivalue = []string{
	`CREATE TABLE svsignal_ivalue (
	id int(11) NOT NULL,
	signal_id int(11) NOT NULL,
	utime int(11) NOT NULL,
	value int(11) NOT NULL,
	offline tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
	`ALTER TABLE svsignal_ivalue
	ADD PRIMARY KEY (id),
	ADD KEY signal_id (signal_id);`,
	`ALTER TABLE svsignal_ivalue
	MODIFY id int(11) NOT NULL AUTO_INCREMENT;`,
	`ALTER TABLE svsignal_ivalue
	ADD CONSTRAINT svsignal_ivalue_ibfk_1 FOREIGN KEY (signal_id) REFERENCES svsignal_signal (id);`,
}

var sql_create_svsignal_mvalue = []string{
	`CREATE TABLE svsignal_mvalue (
	id int(11) NOT NULL,
	signal_id int(11) NOT NULL,
	utime int(11) NOT NULL,
	max double NOT NULL,
	min double NOT NULL,
	mean double NOT NULL,
	median double NOT NULL,
	offline tinyint(1) NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
	`ALTER TABLE svsignal_mvalue
	ADD KEY signal_id (signal_id);`,
	`ALTER TABLE svsignal_mvalue
	ADD CONSTRAINT svsignal_mvalue_ibfk_1 FOREIGN KEY (signal_id) REFERENCES svsignal_signal (id);`,
}

var sql_create_svsignal_signal = []string{
	`CREATE TABLE svsignal_signal (
	id int(11) NOT NULL,
	group_id int(11) NOT NULL,
	signal_key varchar(200) NOT NULL,
	name varchar(200) NOT NULL,
	type_save int(11) NOT NULL,
	period int(11) NOT NULL,
	delta float NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,

	`ALTER TABLE svsignal_signal
	ADD PRIMARY KEY (id),
	ADD KEY group_id (group_id);`,

	`ALTER TABLE svsignal_signal
	MODIFY id int(11) NOT NULL AUTO_INCREMENT;`,

	`ALTER TABLE svsignal_signal
	ADD CONSTRAINT svsignal_signal_ibfk_1 FOREIGN KEY (group_id) REFERENCES svsignal_group (id);`,
}

var sql_create_svsignal_tag = []string{
	`CREATE TABLE svsignal_tag (
	id int(11) NOT NULL,
	signal_id int(11) NOT NULL,
	tag varchar(200) NOT NULL,
	value varchar(400) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
	`ALTER TABLE svsignal_tag ADD PRIMARY KEY (id), ADD KEY signal_id (signal_id);`,
	`ALTER TABLE svsignal_tag MODIFY id int(11) NOT NULL AUTO_INCREMENT;`,
	`ALTER TABLE svsignal_tag ADD CONSTRAINT svsignal_tag_ibfk_1 FOREIGN KEY (signal_id) REFERENCES svsignal_signal (id);`,
}

var sql_create_svsignal_group = []string{
	`CREATE TABLE svsignal_group (
	id int(11) NOT NULL,
	group_key varchar(200) NOT NULL,
	name varchar(200) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
	`ALTER TABLE svsignal_group
ADD PRIMARY KEY (id);`,
	`ALTER TABLE svsignal_group
MODIFY id int(11) NOT NULL AUTO_INCREMENT;`,
}

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func migratedb(db *sql.DB) error {
	fmt.Println("Migrate check Table")
	list_table := []struct {
		name_table string
		sql        []string
	}{
		{name_table: "svsignal_group", sql: sql_create_svsignal_group},
		{name_table: "svsignal_signal", sql: sql_create_svsignal_signal},
		{name_table: "svsignal_fvalue", sql: sql_create_svsignal_fvalue},
		{name_table: "svsignal_ivalue", sql: sql_create_svsignal_ivalue},
		{name_table: "svsignal_mvalue", sql: sql_create_svsignal_mvalue},
		{name_table: "svsignal_tag", sql: sql_create_svsignal_tag},
	}
	/*
		{"svsignal_group", ""},
		{"svsignal_signal", ""},
		{"svsignal_tag", ""},
	*/

	rows, err := db.Query("SHOW TABLES LIKE 'svsignal_%'")
	if err != nil {
		return err
	}
	defer rows.Close()

	list_table_db := []string{}
	for rows.Next() {
		var name_table string
		err := rows.Scan(&name_table)
		if err != nil {
			fmt.Println(err)
			continue
		}
		list_table_db = append(list_table_db, name_table)
	}
	for _, action := range list_table {
		if Contains(list_table_db, action.name_table) {
			fmt.Printf("table:%s is Ok\n", action.name_table)
		} else {
			fmt.Printf("create table:%s\n", action.name_table)
			is_ok := false
			for _, sql := range action.sql {
				_, err := db.Exec(sql)
				if err != nil {
					fmt.Printf("error %s", err)
				} else {
					is_ok = true
				}
			}
			if is_ok {
				fmt.Println("is Ok")
			}
		}
	}

	return nil
}
