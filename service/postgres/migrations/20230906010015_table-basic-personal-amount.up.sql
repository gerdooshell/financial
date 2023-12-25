create table tax.basic_personal_amount (
    id bigint generated always as identity primary key,
    region_code varchar(20) not null,
    year int not null,
    min_amount numeric not null,
    min_amount_salary numeric not null,
    max_amount numeric not null,
    max_amount_salary numeric not null,
    constraint basic_personal_amount_item unique (region_code, year)
)
