--d:postgres

REVOKE ALL ON DATABASE tasker FROM PUBLIC ;
GRANT CONNECT ON DATABASE tasker TO tasker_user ;

GRANT ALL ON FUNCTION activity_is_parent_of ( INTEGER, INTEGER ) TO tasker_user ;

REVOKE ALL ON FUNCTION activity_is_parent_of ( INTEGER, INTEGER ) FROM public ;
