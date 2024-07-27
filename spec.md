# Specification of patch semantic for Protocol Buffers

In general 4 kinds of basic patch operations for arbitrary Protocol Buffer message are distinguished. Those are `set`, `append`, `insert` and `clear`. They operate on a single path.

Additionally, for convenience, 2 special operations were introduced. Those are `copy`, `move` and `swap`. They operate on two paths at the same time.

Although operations relate to any value that can be found within a Protocol Buffer message, basic behavior of a given operation depends on the type of the closes ancestor composite type to the given value. Examples of the closes composite type ancestors:

- For scalar value (`string`, `inst32`, `uint64`, etc.) field, it is a message congaing this field.
- For message value field, it is a message congaing this field.
- For repeated value field (list field), it is a message congaing this field.
- For map value field, it is a message congaing this field.
- For item at specific index in a repeated field (list field), it is a list containing this item.
- For item at specific map key in a map field, it is a map containing this item.

Therefore behavior of all operations for all cases can be mostly described by specifying what operation should do for each type of ancestor composite type, that is messages, list or maps. The only exception for this rule are base messages, that some patch operation is performed directly on, since base messages have no ancestors.

## Set operation

Set operation is the most basic. As the name implies it sets the value of message field, list index or map key specified by a path, known as a **initial value**, to a given value, known as a **replacement value**. The entity containing the initial value is known as a **target element**.

When target element is a filed within some message, set operation must change the field value to the provided replacement value. Analogically, when target element is an index within list or a key within map, the value must be changed to the replacement value. When a message field or a list index is not found within its parent the operation should fail. This is important distinction from set operation semantic in JSON patch. However to make Protocol Buffers patch semantic of maps similar to JSON patch implementation should allow setting nonexisting map keys.

When initial value is a base messages, set operation must set and clear all fields accordingly to the contents of the replacement value, such that the result will be the same as for any message within.

## Append operation

Append operation adds a given value, known as **new value**, to the end of the list, known as **target element** or **target list**, specified by a path. The operation must not be valid when target element is a message field, a map key or a base message.

## Insert operation

Insert operation adds a given value, known as **new value**, at some location within a list, known as **target element** or **target list**, as specified by a path. When the path points to an existing index within the list, operation should insert the provided value at the specified index expanding the list. When the path points to an existing index right outside the list (index is equal to list length) insert operation should behave just like append operation.

The operation must not be valid when target element is a message field, a map key, a base message or the index is out of bounds (except for the case described above, where insert becomes effectively append operation).

## Clear operation

Clear operation clears the field, known as **target element** according to Protocol Buffer rules.

When target element is a message field it means setting field to its zero value. However, for a message field that is part of a oneof, the value should be removed (if it was set within the oneof) as if the oneof has no value. When target element is an item under list index, clear operation must remove this index shrinking the list. When target element is an item under map key, clear operation must remove this key from the map. When target element is a base messages, it means clearing all fields within the message.

## Copy operation

Copy operation is a special extension operation that electively allows to set a value, known as **initial value** and contained by an entity known as **target element**, without providing a concrete replacement, but rather an another path where to find the value, known as **replacement value** within the base message. The copy operation semantic must therefore follow the set operation semantic.

## Move operation

Move operation, similarly to copy operation, is a special extension operation. It replaces an **initial value** contained by **target element** with **replacement value** contained by **replacement element**. The it deletes **replacement element**. The move operation semantic must therefore follow set and delete operation semantics.

## Swap operation

Swap operation, similarly to copy and move operation, is a special extension operation. It electively acts like two set operations that exchanges two values - one known as **first value** and contained by **first element**, with another known as **second value** and contained by **second element**. Therefore swap operation semantic must follow the set operation semantic.
