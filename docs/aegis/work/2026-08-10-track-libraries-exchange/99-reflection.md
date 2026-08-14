# Reflection

Die gewählte Produktgrenze ist tragfähig: RailKeeper besitzt ein eigenes, streng validiertes und
versioniertes Austauschformat, ohne unzuverlässige Unterstützung proprietärer Formate
vorzutäuschen. Der Draft- und Prüfworkflow hält externe Daten von der Planerpalette fern, bis ein
Admin Quelle und Geometrie dokumentiert geprüft hat.

Geometriesnapshots lösen die wichtigste Langzeitinvariante. Eine Bibliothek kann stillgelegt oder
durch eine neue Version ergänzt werden, ohne Zeichnung, Analyse, Diff oder Stückliste bestehender
Pläne nachträglich zu verändern.

Die reale Browserabnahme war entscheidend. Komponenten- und Repositorytests allein hätten weder
die instabile React-Abhängigkeit noch die bisher hartcodierte Tillig-Beschriftung in allen
Planeransichten sichtbar gemacht.
