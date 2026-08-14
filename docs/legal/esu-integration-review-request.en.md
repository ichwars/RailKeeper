# Draft: Request for review of RailKeeper's limited ECoS integration

> **Status: Draft, not sent.**
>
> This letter must not be sent without new explicit approval.

To<br>
ESU electronic solutions ulm GmbH & Co. KG<br>
Edisonallee 29<br>
89231 Neu-Ulm<br>
Germany

## Subject

Request for review of a limited ECoS integration and the function-key symbols used by the free
RailKeeper project

Dear Sir or Madam,

I develop RailKeeper, a free, locally operated, open-source application for managing model railway
vehicles. The project is publicly available at <https://github.com/ichwars/RailKeeper> and is
released under the GNU Affero General Public License Version 3 only (`AGPL-3.0-only`).

RailKeeper is an independent project and is not affiliated with ESU. It does not claim
certification, approval, endorsement, or official compatibility by ESU. ECoS is a trademark of
ESU electronic solutions ulm GmbH & Co. KG.

I would like to ask for your early and transparent review of the deliberately limited integration.
Its sole purpose is to transfer selected locomotive master data, static function-key descriptions,
and CV values into a local vehicle collection. RailKeeper presents a preview and requires explicit
confirmation before applying a change.

## Implemented scope

RailKeeper currently reads only the following information:

- ECoS system object 1: `info` and `status` for connection detection
- locomotive object manager 10: `addr`, `name`, and `protocol` for listing
- individual locomotive objects: `profile`, `protocol`, `name`, `addr`, and `funcdesc`
- additional static decoder information: `cv`, `cvs`, `cvlist`, and `functionmapping`
- targeted CV queries: CV 1 through 8, as well as CV 7, 8, 17, 18, and 29

After a preview and explicit confirmation, RailKeeper can write only `name`, `addr`, and `protocol`
to a selected locomotive object. No other write or control commands are intended.

The following capabilities are deliberately excluded:

- current speed or speed step
- current direction
- active function states or `funcset`
- drive, function, STOP, or GO commands
- switches or magnetic accessories and their object managers
- routes
- S88 or feedback objects
- boosters and their object managers
- all other ECoS object managers
- locomotive images, image references, or image fields from the ECoS

## Function-key symbols

RailKeeper contains locally embedded SVG function-key symbols and mappings. The repository metadata
identifies the following packages and document as their origin or working basis:

- `Funktionstasten_SVG_Variante_1_172_Symbole.zip`
- `ESU_Funktionssymbole_V1_Variante2_Feinlinien_aktiv_inaktiv_SVG.zip`
- `50200_ECoS_Uebersicht_Funktionstastensymbole_ESU_KG_DE-EN_Auflage-3-1.pdf`

These symbols are not downloaded automatically from an ECoS device. I would specifically like to
clarify whether ESU objects to including and distributing these symbols or mappings in the public
RailKeeper repository, and whether special attribution, conditions, or separate permission is
required. The project's AGPL license is not intended to claim or transfer any rights in third-party
trademarks, graphics, documentation, or other third-party rights.

## Request for feedback

Please let me know as specifically as possible:

- whether ESU objects to the described read or write scope,
- whether any individual queries, fields, or write operations should be removed or changed,
- whether the function-key symbols or mappings must be removed, replaced, or separately licensed,
- which trademark, copyright, or other notices ESU considers necessary.

If ESU raises concerns, I am willing to modify or remove the affected functions, data, or graphics
from RailKeeper. Identifying the specific elements in question would allow me to address them
quickly and precisely.

Thank you for your review.

Yours faithfully,

Daniel Roth<br>
GitHub: <https://github.com/ichwars><br>
Project: <https://github.com/ichwars/RailKeeper>
