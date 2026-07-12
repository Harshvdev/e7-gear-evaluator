# Priority of Stats:

### Step 1: Standard Conversion Factors

To normalize different stats (like Speed vs. Attack), they must be converted using standard maximum roll ratios:

- **Attack %, Health %, Defense %, Effectiveness, Effect Resistance:** $\times 1$
    
- **Crit Chance:** $\times 1.6$
    
- **Crit Damage:** $\times 1.14$
    
- **Speed:** $\times 2$
    

### Step 2: Calculate Substat Contribution (GSP)

For flat stats (Attack, HP, Defense), you must first convert the target bonus stat into a percentage based on the hero's base stats:

$$\text{Bonus } \% = \left( \frac{\text{Target Stat} - \text{Base Stat}}{\text{Base Stat}} \right) \times 100$$

Once everything is in percentages or normalized values, calculate the GSP for each stat $i$:

$$\text{GSP}_i = \text{Bonus Value}_i \times \text{Conversion Factor}_i$$

## Walkthrough: Mathematically Profiling Afternoon Soak Flan

Using your target build and her base stats at Level 60 (Awakened):

- **Base Stats:** Attack = 1222, HP = 5784, Defense = 652, Speed = 115, Crit CC = 15%, Crit CD = 150%.
    
- **Target Stats:** Attack = 4440, HP = 13113, Defense = 1062, Speed = 183, CC = 39%, CD = 347%.
    

### 1. Calculate the GSP for each stat:

- **Attack GSP:**
    
    $$\text{Bonus } \% = \left( \frac{4440 - 1222}{1222} \right) \times 100 = 263.3\%$$
    
    $$\text{GSP}_{\text{Atk}} = 263.3 \times 1 = \mathbf{263.3}$$
    
- **Crit Damage GSP:**
    
    $$\text{Bonus Value} = 347\% - 150\% = 197\%$$
    
    $$\text{GSP}_{\text{CD}} = 197 \times 1.14 = \mathbf{224.6}$$
    
- **Health GSP:**
    
    $$\text{Bonus } \% = \left( \frac{13113 - 5784}{5784} \right) \times 100 = 126.7\%$$
    
    $$\text{GSP}_{\text{HP}} = 126.7 \times 1 = \mathbf{126.7}$$
    
- **Speed GSP:**
    
    $$\text{Bonus Value} = 183 - 115 = 68 \text{ Speed}$$
    
    $$\text{GSP}_{\text{Spd}} = 68 \times 2 = \mathbf{136.0}$$
    
- **Defense GSP:**
    
    $$\text{Bonus } \% = \left( \frac{1062 - 652}{652} \right) \times 100 = 62.8\%$$
    
    $$\text{GSP}_{\text{Def}} = 62.8 \times 1 = \mathbf{62.8}$$
    
- **Crit Chance GSP:**
    
    $$\text{Bonus Value} = 39\% - 15\% = 24\%$$
    
    $$\text{GSP}_{\text{CC}} = 24 \times 1.6 = \mathbf{38.4}$$

### 2. The Final Priority Matrix

Add all GSPs together to find the Total Gear Budget ($\text{Total GSP} = 851.8$). Divide each individual GSP by the total to find the exact mathematical priority:

|**Stat**|**GSP Consumed**|**Budget % (Priority Weight)**|**Mathematical Ranking**|
|---|---|---|---|
|**Attack**|263.3|**30.9%**|**Priority 1 (Primary)**|
|**Crit Damage**|224.6|**26.4%**|**Priority 2 (Primary)**|
|**Speed**|136.0|**16.0%**|**Priority 3 (Secondary)**|
|**Health**|126.7|**14.9%**|**Priority 4 (Secondary)**|
|**Defense**|62.8|**7.4%**|**Priority 5 (Filler)**|
|**Crit Chance**|38.4|**4.5%**|**Priority 6 (Filler)**|

---
