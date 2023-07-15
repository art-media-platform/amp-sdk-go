@0x9aff325096b39f47;

using Go = import "/go.capnp";
$Go.package("arc");
$Go.import("github.com/arcspace/go-arc-sdk");

using CSharp = import "/csharp.capnp";
$CSharp.namespace("Arcspace");

struct AttrDefTest {
    typeName @0 :Text;
    typeID   @1 :Int32;   
}

const linkCellSpec        :Text = "(CellInfo)()";
#const linkCellSpec        :Text = "(CellInfo,[Locale.Name]Labels,[Purpose.Name]AssetRef:Glyphs[Surface.Name]Positions)()";




# testing
const builtInDefs :List(AttrDefTest) = [
    (typeName = "arc.CellInfo", typeID = 123),
    (typeName = "arc.GeoFix",   typeID = 456),
];



enum URIScheme2 {
    any @0;
    data @1;
    http @2;
    file @3;
    # crateAsset @4;
    # cellSchema @5;
}

struct AssetRef2 {

    label           @0 :Text;
    # Describes the asset (optional)
    
    mediaType       @1 :Text;
    # Describes content of URI; MIME type (or '/' separated type pathname)
    
    scheme          @2 :URIScheme2;
    # Describes URI scheme such that the pa URL scheme is not required to prefix URI
    
    uri             @3 :Text;
    # URI to the asset (has scheme prefix if URIScheme == URL, otherwise, scheme prefix is optional)
    
    attrs           @4 :Attrs;
    # Open-ended meta data
    
    struct Attrs {
        pixWidth       @0 :Int32;
        pixHeight      @1 :Int32;
        physWidth      @2 :Float32;
        physHeight     @3 :Float32;
    }
}


