CREATE TABLE companies (
    name        text,
	inn         text,
	phone       text,
	address     text,
	individual  text
)

ALTER TABLE companies ADD CONSTRAINT unique_name UNIQUE(name);
ALTER TABLE companies ADD CONSTRAINT unique_inn UNIQUE(inn);
