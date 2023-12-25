create or replace function tax.get_max_contributory_earning(p_year numeric) returns numeric
    language plpgsql
as
$$
    DECLARE
row record;
BEGIN
select * into row from tax.canada_pension_plan cpp where cpp.year = p_year;
return row.maximum_annual_pensionable_earning - row.basic_exception_amount;
END;$$;