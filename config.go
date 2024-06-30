package main

const ConfigTable = `
CREATE TABLE IF NOT EXISTS config (
 name KEY NOT NULL, 
 service VARCHAR(8) NOT NULL COLLATE NOCASE,
 value NOT NULL,
 desc TEXT NOT NULL COLLATE NOCASE
);
`

type Config struct {
	name    string
	service string
	value   any
	desc    string
}

func CreateConfig(name string, service string, value any, desc string) error {
	c := Config{name, service, value, desc}
	_, err := DB.ExecContext(ServicesContext, "INSERT OR REPLACE INTO config VALUES (?,?,?,?)",
		c.name, c.service, c.value, c.desc)
	return err
}
