CREATE DATABASE project;
USE  project;
create table `firmaTemplates` (
`Name` varchar(64) not null,
primary key (`Name`)
) comment 'Firmas';
create table `firma` (
`Name` varchar(64) not null,
`FirmaCode` tinyint not null auto_increment unique,
`Descripton` varchar(64) default "",
primary Key (`FirmaCode`) 
)comment 'Firmas';
create table `Default_Firma` (
`Name` varchar(64) not null unique
) comment 'Firma';
create table `Default_Table` (
`Name` varchar(64) not null unique
) comment 'Table';
create table `tableTemplates` like `Default_Firma`;
SET SQL_SAFE_UPDATES = 0;
/*////////////////////////////////////////////////////////////////////////


*/
show tables;

create table `firma`(
`Id` int not null auto_increment unique,
`Name` varchar(64) not null,
primary key (`id`)
);

create table `tabel` (
`Id` int not null,
`Name` varchar(64) not null,
`firma_id` int not null,
`property` varchar(64) not null,
`value` varchar(255),
primary key (`id`)
);

create table `Data` (
`property` varchar(64) not null,
`Value` varchar(255),
`id` int not null,
`tabel_id` int not null
);
drop table `data`;

insert into `Data` (Name , Value , id,parent_id) values ('Name', 'Nikolaev','2','0');
select * from data;





