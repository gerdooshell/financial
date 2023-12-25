create or replace function tax.get_max_annual_contribution(p_year numeric) returns numeric
    LANGUAGE plpgsql
as
$$
    DECLARE
row RECORD;
BEGIN
select * into row from tax.canada_pension_plan cpp where cpp.year = p_year;
return (row.maximum_annual_pensionable_earning - row.basic_exception_amount) * row.rate / 100;
END;
$$;
