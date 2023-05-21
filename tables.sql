CREATE TABLE IF NOT EXISTS authors (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS blogs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    author INT,
    content TEXT,
    image JSON,
    FOREIGN KEY (author) REFERENCES authors(id),
    FULLTEXT (content)
);

INSERT INTO authors(name) VALUES ("John");
INSERT INTO authors(name) VALUES ("Jim");