An API backend that integrates a custom IMAP client in order to receive, categorize, and organize newsletters.

# Prerequisites

In order to run this backend, you need to set up:

- a private mail server. You only need to be able to receive mail, which makes the process relatively simpler. I set one up on EC2 using Postfix as the MTA and Dovecot as the MDA.
- a MongoDB instance to hold API data.
- a MySQL instance for Dovecot SASL.

# Usage

Once you have all the prerequisites listed above, set the following environment variables:

```bash
$ export ALEPH_ENV=your environment # default: development
$ export DOMAIN=your-mail-server-domain
$ export MONGO_URI=your-mongo-uri
$ export MYSQL_URI=your-my-sql-uri
```

Now you can run backend as follows, from the project root directory:

```bash
$ go run aleph/backend
```