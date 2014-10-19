Hierarchical Temporal Memory implementation.

Inpired by htm from zacg, which in turn is a port of nupic by Numenta. This is, however, a full rewrite.

https://github.com/numenta/nupic/wiki/Contributing-to-NuPIC

Design Doc (reverse engineered)
-------------------------------

Objectives of the CLA modeling:
1. give similar things similar representations (spatial pooling)
2. learn temporal similarities (temporal pooling)

A column is connected to around 50% of the bits of an input. Each connection has a permanence weight which increases as the synapses strenghten and decreases as they weaken. Above a certain permanence threshold, we say they are connected. Below that, they are disconnected.

At each step, we look at the bits of the input which correspond to connected synapses (potential pool) and count how many are set (overlap score). The top 2% scoring cells then "fire", which means they are now activated.

When learning, the column then updates the permanence weights to better match the input. The actual increment and decrement values are swarmed over (i.e. trained).

A column has a sequence of cells that can be in one of three states: INACTIVE, ACTIVE, PREDICTED. When a column fires and it has a predicted bit set, these bits become active. When the predicted set is empty, it means we are seeing a new pattern, so we activate all cells (bursting). We can then learn a transition from the previous state to the new state by randomly selecting one of the active cells to predict its activation next time the previous state happens.

The usefulness of bursting is in anomaly detection: the ratio of bursting columns directly reflects how different this input is from predictable inputs that the system has seen before.

### Data flow

* Sensor region: consumes a dictionary input and converts the real world data into a sparse binary representation.
* Spatial pool region: picks which columns are active. Since 2% of columns are active, for a pool of 2048 columns, around 40 are active. So the output of the spatial pool is a vector of around 40 integers representing the active columns. Note that this is wildly paralelizable.
* Temporal pool region: picks the cells in a column that are active. A TP is typically 32 cells long, so the output of the Temporal pool is again 2048 32-bit integers (sparse, only 40 integers are not zero).
* Classifier: keeps a histogram of output value per input value, tring to predict which value will appear two steps ahead.
    For scalars, the classifier also keeps a granularization table: each bucket of discretized values from the sensor region keeps a moving average of values that appeared for that bucket.

### Potential pool

The potential pool is the set of all inputs that are within a column's reach. There are no connections between a column and inputs outside its potential pool.

The potential pool can be represented as a per-column map of input index to permanence weight. Since a column is typically connected to a small subset of the inputs, the map is mostly empty. The original system stores binary connections, which has a tradeoff of a smaller map to iterate when computing each step, but still needs the permanence weights for learning.

Two major operations take place in the potential pool. One is determining the overlap between an input and the connected columns (i.e. which columns that have a permanence above a threshold T). Because inputs are stored as Bitsets, this can be done efficiently if the connected columns are represented as a Bitset as well, so we can perform an AND operation between the two. If memory allocation for the indices vector is done carefully, this could work very well.

The second operation is learning, which would require going through the input AND NOT connected for increments, and connected AND NOT input for decrements. Might be that iterating on the input once and checking bits individually works better.

### Region
A region is a named set of one or more identical nodes. Multiple regions can belong to a network.

- A node within a region is represented by a cell in an n-dimensional grid, identified by a coordinate X = {x1, x2, x3, x4, ..., xN}.
- A Dimensions object has the size of each dimension. A 3-dimensional grid of 4x3x3 would be described as D = [4, 3, 3], and a valid coordinates would range within
 {[0, 4), [0, 3), [0, 3)}.
- A region must map from coordinate to index. For our 4,3,3 example, a valid coordinate is 3, 0, 2, which translates to 3*4*3*3 + 0*3*3 + 2*3. Maybe a more interesting byte-aligned index representation would work just as well but be faster to translate back?
- A region has inputs and outputs. More below.
- A region has a type of TestNode, VectorFileEffector, VectorFileSensor, or a dynamic python type. I think I'll go ahead and redesign this bit. More below.

- There are all sorts of get/set parameters, serialization and abstraction leaks that will be added as needed, if needed.

### Inputs and Outputs

Inputs and Outputs are named.

### Encoders
https://www.youtube.com/watch?v=3gjVVNPnPYA&feature=youtu.be&t=2m40s

Encoders take real world input (e.g. "4" or "orange") and turn them into a sparse representation. The representation is as an array of N bits, of which W bits are on and the rest is off. W is much smaller than N, so it is sparse.

The creation of the encoders is usually an offline process and Numenta swarms over N, usually in the range 28 to 521, according to the video. For categories, they usually take W random bits for each category (must be uniform to keep the entropy). For scalars, they look at the range of possible values, e.g [0.0, 114.0] and project that into buckets of W consecutive bits, so 0.0 becomes 000000...111 and 114.0 becomes 111...000 and the middle value has the middle bits on. This has nice properties but has just as much entropy as the categories. Good job.

### Parsing

- CSV format works well; some other tabular file format would also work.
