create table tax.canada_pension_plan (
    id bigint generated always as identity primary key,
    year int not null unique,
    maximum_annual_pensionable_earning numeric not null,
    basic_exception_amount numeric not null,
    rate numeric not null
);

create unique index if not exists cpp_year on tax.canada_pension_plan (year desc);
