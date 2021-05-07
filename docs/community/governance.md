---
title: "Governance"
---

# Governance

This document describes the rules and governance of the project. It is a slightly modified version of the [Prometheus Governance](https://prometheus.io/governance/#governance-changes) and similar to the [Thanos Governance](https://thanos.io/tip/thanos/governance.md/)

It is meant to be followed by all the developers of the Observatorium project and the Observatorium community. Common terminology used in this governance document are listed below:

* **Maintainers Team**: A core Observatorium team that have owner access to the https://github.com/observatorium organization and all projects within it. Current list is available [here][maintainers-doc].

* **The Observatorium project**: The sum of all activities performed under the [Observatorium organization on GitHub][gh], concerning one or more repositories or the community.

Maintainers are part of [`observatorium-team@googlegroups.com`][team] email list.

## Values

The Observatorium developers and community are expected to follow the values defined in the [CNCF charter][charter], including the [CNCF Code of Conduct][coc].
Furthermore, the Observatorium community strives for kindness, giving feedback effectively, and building a welcoming environment. The Observatorium developers generally decide by consensus and only resort to conflict resolution by a majority vote if consensus cannot be reached.

## Decision making

### Maintainers Team

Team member status may be given to those who have made ongoing contributions to the Observatorium project for at least 3 months.
This is usually in the form of code improvements and/or notable work on documentation, but organizing events or user support could also be taken into account.

New members may be proposed by any existing Maintainer by sending an email to the [Observatorium][team]. It is highly desirable to reach consensus about acceptance of a new member.
However, the proposal is ultimately voted on by a formal [supermajority vote](#supermajority-vote) of Team Maintainers.

If the new member proposal is accepted, the proposed team member should be contacted privately via email to confirm or deny their acceptance of team membership.
This email will also be CC'd to the [Observatorium][team] for record-keeping purposes.

If they choose to accept, the following steps are taken:

* Maintainer is added to the [GitHub organization][gh] as an _Owner_.
* Maintainer is added to the [Observatorium][team].
* Maintainer is added to the list of team members [here][maintainers-doc]
* New maintainer is announced on the [Observatorium Twitter](https://twitter.com/0bservatorium) by an existing team member.

Team members may retire at any time by emailing the [observatorium-team@googlegroups.com][team] mailing list.

Team members can be removed by [supermajority vote](#supermajority-vote) on the [observatorium-team@googlegroups.com][team] mailing list. For this vote, the member in question is not eligible to vote and does not count towards the quorum.

Upon death of a member, their team membership ends automatically.

### Technical decisions

Smaller technical decisions are made informally and [lazy consensus](#consensus) is assumed. Technical decisions that span multiple parts of the Observatorium project
should be discussed and finalized on [GitHub issues][issues] and in most cases should be followed by a proposal as described [here](/CONTRIBUTING.md#adding-new-features--components).

Decisions are usually made by [lazy consensus](#consensus). If no consensus can be reached, the matter may be resolved by [majority vote](#majority-vote).

### Governance changes

Material changes to this document are discussed publicly on the [Observatorium GitHub](https://github.com/observatorium/observatorium) repository.
Any change requires a [supermajority](#supermajority-vote) in favor. Editorial changes may be made by [lazy consensus](#consensus) unless challenged.

### Other matters

Any matter that needs a decision, including but not limited to financial matters, may be called to a vote by any Maintainer if they deem it necessary.
For financial, private, or personnel matters, discussion and voting takes place on the [observatorium-team@googlegroups.com][team] email list. Otherwise, discussions and votes are held in public on the GitHub issues or #observatorium-dev CNCF slack channel.

## Voting

The Observatorium project usually runs by informal consensus, however sometimes a formal decision must be made.

Depending on the subject matter, as laid out [above](#decision-making), different methods of voting are used.

For all votes, voting must be open for at least one week. The end date should be clearly stated in the call to vote.
A vote may be called and closed early if enough votes have come in one way so that further votes cannot change the final decision.

In all cases, all and only [Maintainers](#maintainers-team) are eligible to vote, with the sole exception of the forced removal of a team member, in which case said member is not eligible to vote.

Discussions and votes on personnel matters (including but not limited to team membership and maintainership) are held in private on the [observatorium-team@googlegroups.com][team] email list. All other discussion and votes are held in public on the GitHub issues or #observatorium-dev CNCF slack channel.

For public discussions, anyone interested is encouraged to participate. Formal power to object or vote is limited to members of the [Maintainers Team](#maintainers-team).

### Governance

It's important for the project to stay independent and focused on shared interests instead of on the single use case of one company or organization.

We value open source values and freedom, that's why we limit Maintainers Team **votes to a maximum of two from a single organization or company.**

We also encourage any other company interested in helping to maintain Observatorium to join us to make sure we stay independent.

### Consensus

The default decision making mechanism for the Observatorium project is [lazy consensus][lazy]. This means that any decision on technical issues is considered supported by the [team][team] as long as nobody objects.

Silence on any consensus decision is implicit agreement and equivalent to explicit agreement. Explicit agreement may be stated at will.

Consensus decisions can never override or go against the spirit of an earlier explicit vote.

If any [member of the Maintainers Team](#maintainers-team) raises objections, the team members work together towards a solution that all involved can accept.
This solution is again subject to lazy consensus.

In case no consensus can be found, but a decision one way or the other must be made, anyone from the [Maintainers Team](#maintainers-team) may call a formal [majority vote](#majority-vote).

### Majority vote

Majority votes must be called explicitly in a separate thread on the appropriate mailing list. The subject must be prefixed with `[VOTE]`.
In the body, the call to vote must state the proposal being voted on. It should reference any discussion leading up to this point.

Votes may take the form of a single proposal, with the option to vote yes or no, or the form of multiple alternatives.

A vote on a single proposal is considered successful if more vote in favor than against.

If there are multiple alternatives, members may vote for one or more alternatives, or vote “no” to object to all alternatives.
It is not possible to cast an “abstain” vote. A vote on multiple alternatives is considered decided in favor of one alternative if it has received the most votes in favor, and a vote from more than half of those voting. Should no alternative reach this quorum, another vote on a reduced number of options may be called separately.

### Supermajority vote

Supermajority votes must be called explicitly in a separate thread on the appropriate mailing list.
The subject must be prefixed with `[VOTE]`. In the body, the call to vote must state the proposal being voted on. It should reference any discussion leading up to this point.

Votes may take the form of a single proposal, with the option to vote yes or no, or the form of multiple alternatives.

A vote on a single proposal is considered successful if at least two thirds of those eligible to vote vote in favor.

If there are multiple alternatives, members may vote for one or more alternatives, or vote “no” to object to all alternatives.
A vote on multiple alternatives is considered decided in favor of one alternative if it has received the most votes in favor, and a vote from at least two thirds of those eligible to vote. Should no alternative reach this quorum, another vote on a reduced number of options may be called separately.

## FAQ

This section is informational. In case of disagreement, the rules above overrule any FAQ.

### For a majority vote, what if there is an even number of maintainers and an equal amount of votes in favor and against?

It has to be majority so the vote will be declined.

### So what's the TLDR difference between majority vs supermajority?

It's about the number of up votes needed to agree on the decision.

* majority: a majority of voters has to agree.
* supermajority: 2/3 of voters have to agree.

### How do I propose a decision?

See [Contributor doc](/CONTRIBUTING.md#adding-new-features--components)

### How do I become a team member?

To become an official member of the Maintainers Team, you should make ongoing contributions to one or more project(s) for at least three months.
At that point, a team member (typically a maintainer of the project) may propose you for membership.
The discussion about this will be held in private, and you will be informed privately when a decision has been made. A possible, but not required, graduation path is to become a triage member first.

Should the decision be in favor, your new membership will also be announced on the [Observatorium Twitter][twitter].

### How do I add a project?

As a team member, propose the new project on the [Observatorium GitHub Issues][issues]. However, currently to maintain a project in our organization you have to first become an Observatorium Maintainer.

All are encourage to start their own project related to Observatorium. The Observatorium team is happy to link to your project in appropriate page.

### How do I remove a Maintainer member?

All members may resign by notifying the [observatorium-team@googlegroups.com][team]. If you think a team member should be removed against their will, propose this to the [observatorium-team@googlegroups.com][team] mailing list.
Discussions will be held there in private.

### Can a majority/supermajority vote be held in a GitHub PR by just approving the PR?

No, a `[VOTE]` email has to be created.

### What if during a majority/supermajority vote there is no answer after week?

For majority votes this means that members who did not send a response agree with the proposal.

For supermajority votes the team has to wait for all answers.

[twitter]: https://twitter.com/0bservatorium
[issues]: https://github.com/observatorium/observatorium/issues
[maintainers-doc]: https://observatorium.io/docs/community/observatorium/maintainers.md/
[team]: https://groups.google.com/forum/#!forum/observatorium-team
[gh]: https://github.com/observatorium
[charter]: https://www.cncf.io/about/charter/
[coc]: https://github.com/cncf/foundation/blob/master/code-of-conduct.md
[lazy]: https://couchdb.apache.org/bylaws.html#lazy
