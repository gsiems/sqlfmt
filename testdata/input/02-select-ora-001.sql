--d:oracle

-- A fairly short query with joins, a comment, and a few parens
SELECT exp.line_num,
        exp.other_line_num,
        row_number () OVER (
            PARTITION BY pol.po_id, exp.other_line_num
            ORDER BY exp.p_dt, exp.line_num, exp.invoice_id, exp.monetary_amount ) AS line_num,
        pol.po_id,
        'PAY' AS cost_code,
        exp.p_dt AS payment_date,
        exp.invoice_id,
        round (
            CASE
                WHEN pol.po_type IN ( 'A', 'B', 'C' ) THEN exp.merchandise_amount
                ELSE exp.monetary_amount
                END, 2 ) AS total_approved_amt,
        bcat.category_code
    FROM schema1.po_line pol
    JOIN schema1.expenditure exp
        ON ( exp.po_id = pol.po_id
            AND exp.line_num = pol.line_num )
    JOIN schema2.budget_category bcat
        ON ( bcat.category_id = pol.category_id )
    WHERE exp.monetary_amount <> 0 ;
