# blind-turn

Never made a turn server before. This one will be pseudonymous. Hence the name,
"Blind Turn."

Now that it's not late on Sunday night, an explanation. Using the power of Go
Interfaces and SAM, it is possible to adapt any application to the use of I2P
for traffic. That includes STUN/TURN servers like this one. Since pion.ly makes
these interfaces part of how you may configure their library, you can actually
substitute them out directly to create a pure-I2P STUN/TURN server.

## Why is it useful?

- Normally, peer-to-peer connections leak a great deal of metadata to:
 1. Participants in the peer-to-peer interaction.
 2. Network observers who can see the interaction.
 using I2P allows us to avoid this issue by introducing hops between the
 connection which are agnostic of the participant, while preserving the
 essentially peer-to-peer topology.
- Obviously media devices pose an anonymity risk by providing access to
 environmental information from around the user's computer, including the
 possibility of camera images.
 1. You don't have to use them. That's the whole point of the fine work being
  done by every remotely responsible browser vendor, i.e. requiring permissions
  for browser features that may leak on a per-domain basis. TL:DR this is why we
  have containers.
- Anonymity and Pseudonimity provide us with *choices* about what information
 we reveal which are **never** available if anonymity is not. In the case of
 transferring media peer-to-peer, it seems perfectly reasonable in many threat
 models to strive for anonymity from the network, and from the STUN/TURN
 server, but perhaps only wish to hide one's physical location form a peer.

Many, many thanks to pion.ly for their incredible, easy to use library.