create or replace function tax.get_annual_contribution(p_year numeric, p_earning numeric) returns numeric
    LANGUAGE plpgsql
as
$$
    DECLARE
row RECORD;
BEGIN
select * into row from tax.canada_pension_plan cpp where cpp.year = p_year;
if p_earning > row.maximum_annual_pensionable_earning then
            return (row.maximum_annual_pensionable_earning - row.basic_exception_amount) * row.rate / 100;
end if;
        if p_earning <= row.basic_exception_amount then
            return 0;
end if;
return (p_earning - row.basic_exception_amount) * row.rate / 100;
END;
$$;
