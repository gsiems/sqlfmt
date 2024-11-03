--d:postgres

REVOKE ALL ON DATABASE tasker FROM PUBLIC ;
GRANT CONNECT ON DATABASE tasker TO tasker_user ;


GRANT ALL ON FUNCTION activity_is_parent_of ( integer, integer ) TO tasker_user ;

REVOKE ALL ON FUNCTION activity_is_parent_of ( integer, integer ) FROM public ;
