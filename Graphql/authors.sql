create table authors (
	id UUID NOT NULL PRIMARY KEY,
	firstname VARCHAR(50) NOT NULL,
	lastname VARCHAR(50) NOT NULL,
	username VARCHAR(50) NOT NULL,
	password VARCHAR(150) NOT NULL
);
create table articles (
	id UUID NOT NULL PRIMARY KEY,
	author UUID REFERENCES authors(id) NOT NULL ,
	title VARCHAR(50) NOT NULL,
	content TEXT NOT NULL
);
