create table tax.federal_employment_insurance_premium (
    id bigint generated always as identity primary key,
    year int not null unique,
    max_insurable_earning numeric not null,
    employee_rate numeric not null,
    employer_rate numeric not null
)
