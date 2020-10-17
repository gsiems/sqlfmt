package main

type ddl struct {
	state int
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to DDL statements.

ASSERTION: tagging DDL is the final tagging operation, therefore
anything not otherwise tagged is considered to be DDL

*/
func (o *ddl) tag(q *queue) (err error) {
	for i := 0; i < len(q.items); i++ {
		if q.items[i].Type == Unknown {
			q.items[i].Type = DDL
		}
	}
	return err
}
