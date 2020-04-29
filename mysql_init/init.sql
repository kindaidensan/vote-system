create database vote;
create table if not exists vote.agenda(
    id int auto_increment primary key,
    title varchar(255) not null,
    owner varchar(15) not null,
    description varchar(255) not null,
    created_at datetime default current_timestamp,
    closed_at datetime
);
create table if not exists vote.voteduser(
    id int auto_increment primary key,
    agenda_id int,
    user_id varchar(15) not null,
    agree bit(1) not null,
    created_at datetime default current_timestamp
    foreign key (agenda_id) references vote.agenda(id)
);