drop database IF exists `gobatis`;
create database `gobatis`;
use `gobatis`;

create table student
(
    id          int auto_increment primary key,
    name        varchar(20) null,
    age         int         null,
    create_time datetime    null
);

select  exists (select 1 from student where name='') as 'exists'