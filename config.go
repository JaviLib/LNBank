package main

const ConfigTable = `
CREATE TABLE IF NOT EXISTS config (
 name VARCHAR NOT NULL, 
 service VARCHAR NOT NULL ,
 value NOT NULL,
  UNIQUE(name,service)
);
`

type Config struct {
	name    string
	service string
	value   any
}

func CreateConfig(name string, service string, value any) error {
	_, err := DB.ExecContext(ServicesContext, ConfigTable)
	if err != nil {
		return err
	}
	c := Config{name, service, value}
	_, err = DB.ExecContext(ServicesContext, "INSERT OR REPLACE INTO config VALUES (?,?,?)",
		c.name, c.service, c.value)
	return err
}

func SetConfig(name string, service string, value any) error {
	return CreateConfig(name, service, value)
}

func ReadConfig(name string, service string, defaultvalue any) (any, error) {
	row := DB.QueryRowContext(ServicesContext,
		"select value from config where service=? and name=? limit 1", service, name)
	var value any
	if err := row.Scan(&value); err != nil {
		return defaultvalue, CreateConfig(name, service, defaultvalue)
	}
	return value, nil
}
