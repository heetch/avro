package avro

//message EyeballCoord {
//    double latitude = 1;
//    double longitude = 2;
//}
//
//message Eyeball {
//    string id = 1;
//    string user_id = 2;
//    string city_id = 3;
//    EyeballCoord position = 4;
//    EyeballCoord marker_position = 5;
//    string selected_product_id = 6;
//    int64 nearest_driver_eta = 7;
//    int64 nearest_drivers_count = 8;
//    int64 created_at = 9;
//}
//
//message EyeballId {
//    string id = 1;
//}
//

const sample = `
{
                "name": "sample",
                "type": "record",
                "fields": [
                    {
                        "name": "header",
                        "type": [
                            "null",
                            {
                                "name": "Data0",
                                "type": "record",
                                "fields": [
                                    {
                                        "name": "uuid",
                                        "type": [
                                            "null",
                                            {
                                                "name": "UUID0",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "uuid",
                                                        "type": "string",
                                                        "default": ""
                                                    }
                                                ],
                                                "namespace": "headerworks.datatype",
                                                "doc": "A Universally Unique Identifier, in canonical form in lowercase. Example: de305d54-75b4-431b-adb2-eb6b9e546014"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Unique identifier for the event used for de-duplication and tracing."
                                    },
                                    {
                                        "name": "hostname",
                                        "type": [
                                            "null",
                                            "string"
                                        ],
                                        "default": null,
                                        "doc": "Fully qualified name of the host that generated the event that generated the data."
                                    },
                                    {
                                        "name": "trace",
                                        "type": [
                                            "null",
                                            {
                                                "name": "Trace0",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "traceId",
                                                        "type": [
                                                            "null",
                                                            "headerworks.datatype.UUID0"
                                                        ],
                                                        "default": null,
                                                        "doc": "Trace Identifier"
                                                    }
                                                ],
                                                "doc": "Trace0"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Trace information not redundant with this object"
                                    }
                                ],
                                "namespace": "headerworks",
                                "doc": "Common information related to the event which must be included in any clean event"
                            }
                        ],
                        "default": null,
                        "doc": "Core data information required for any event"
                    },
                    {
                        "name": "body",
                        "type": [
                            "null",
                            {
                                "name": "Data1",
                                "type": "record",
                                "fields": [
                                    {
                                        "name": "uuid",
                                        "type": [
                                            "null",
                                            {
                                                "name": "UUID1",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "uuid",
                                                        "type": "string",
                                                        "default": ""
                                                    }
                                                ],
                                                "namespace": "bodyworks.datatype",
                                                "doc": "A Universally Unique Identifier, in canonical form in lowercase. Example: de305d54-75b4-431b-adb2-eb6b9e546014"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Unique identifier for the event used for de-duplication and tracing."
                                    },
                                    {
                                        "name": "hostname",
                                        "type": [
                                            "null",
                                            "string"
                                        ],
                                        "default": null,
                                        "doc": "Fully qualified name of the host that generated the event that generated the data."
                                    },
                                    {
                                        "name": "trace",
                                        "type": [
                                            "null",
                                            {
                                                "name": "Trace1",
                                                "type": "record",
                                                "fields": [
                                                    {
                                                        "name": "traceId",
                                                        "type": [
                                                            "null",
                                                            "headerworks.datatype.UUID0"
                                                        ],
                                                        "default": null,
                                                        "doc": "Trace Identifier"
                                                    }
                                                ],
                                                "doc": "Trace1"
                                            }
                                        ],
                                        "default": null,
                                        "doc": "Trace information not redundant with this object"
                                    }
                                ],
                                "namespace": "bodyworks",
                                "doc": "Common information related to the event which must be included in any clean event"
                            }
                        ],
                        "default": null,
                        "doc": "Core data information required for any event"
                    }
                ],
                "namespace": "com.avro.test",
                "doc": "GoGen test"
            }
`
