# NUDocs
NUDocs is a system for collaborative editing (think Google Docs). Unlike Google Docs, however, it is 100% peer-to-peer. There is no centralized server, and no hierarchy among the peers. It also contains a pluggable interface, so each user can use whichever text editor they prefer, provided that they implement our protocol. In theory, any extensible text editor will work. Plugins have so far been developed for Emacs and the browser.

# Background
Computer supported cooperative work (CSCW) has been around since the late 1980s, prompted by the seminal work [Concurrency Control in Groupware Systems](https://www.lri.fr/~mbl/ENS/CSCW/2017/papers/Ellis-SIGMOD89.pdf). This initial algorithm proposed a peer-to-peer system, however, it was later discovered to be incomplete [COR95][1].

Most modern implementations like Google Docs rely on a centralized server. This offers certain advantages, like being able to create a total ordering of operations that happen on distributed nodes. It has costs as well though, like forcing users to trust a central authority.

The way we chose to get around this issue is by using the REDUCE algorithm, which is a peer-to-peer operational transformation algorithm [SUN98][2].

## Operational Transformation
Operational transformation is a method for maintaining consistency and correctness between different sites running the collaborative editing software. The basic idea is that when an edit is proposed by a user (by them typing it into their editor), the change happens immediately for them, just as if they were typing in a regular text document. Then it sends that that operation to the other peers. The other peers will transform the operation they have received against their state, which may have changed if they made concurrent edits, or received edits from another site earlier. After the operation has been transformed, it is then applied to the document by the peer.

# Demo
![](report/DSFinalDemo.gif "Recorded demo")

# Presentation
View our slides [here](https://docs.google.com/presentation/d/1qCR7XHLZH9ZFzkX1aJn-CQAvKBsPEe_x8KRro4J_kMI/edit?usp=sharing).


[1]: https://cs.uwaterloo.ca/research/tr/1995/08/dopt.pdf
[2]: http://salvin.jeancharles.free.fr/Documents/Projet%20-%20Boulot/NTU-Singapore/p63-sun.pdf