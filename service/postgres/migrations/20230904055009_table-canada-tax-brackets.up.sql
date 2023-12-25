create table tax.canada_tax_brackets (
    id bigint generated always as identity primary key,
    bracket_index int not null,
    region_code varchar(20) not null,
    region_name varchar(100),
    year int not null,
    min_salary numeric not null,
    max_salary numeric not null,
    rate numeric not null,
    constraint canada_tax_brackets_item unique (bracket_index, region_code, year)
    );
