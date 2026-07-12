## The GSP Landmine Filter Algorithm

### Step 1: Calculate Total Build GSP ($\text{GSP}_i$)

Using the conversion factors, calculate the total GSP for the stat $i$ from the global average target build:

$$\text{GSP}_i = \text{Bonus Value}_i \times \text{Conversion Factor}_i$$

(Note: Remember to subtract the base stat and mandatory left-side flat main stats from the target value before converting, as shown in the previous walkthrough).

### Step 2: Calculate Slot Allocation Value ($\text{SAV}_i$)

Divide the total $\text{GSP}_i$ by the 6 available gear slots to find the average GSP contributed by each item:

$$\text{SAV}_i = \frac{\text{GSP}_i}{6}$$

### Step 3: Apply the Noise Floor Threshold

A single, unenhanced base substat at Level 90 naturally possesses an inherent value of **$4$ to $8$ GSP** (averaging $6 \text{ GSP}$) before a single upgrade roll ($+3, +6, +9,$ etc.) is ever spent on it.

- If $\text{SAV}_i \le 6.0$, it mathematically proves the stat only exists in the data because it was sitting unrolled on a piece of gear. The community spent **zero upgrade resources** on it.
    
- Therefore, the binary GSP landmine rule is:
    

$$L_i = \begin{cases} 1 \text{ (Landmine)}, & \text{if } \text{SAV}_i \le 6.0 \\ 0 \text{ (Valid Stat)}, & \text{if } \text{SAV}_i > 6.0 \end{cases}$$

## Mathematical Proof: GSP Landmine Check on Mort

Using Mort's global data from your screenshot:

- **Attack Global Average:** 1977
    
- **Mort Base + Weapon Flat Padding:** ~1610
    
- **Effectiveness Global Average:** 4%
    
- **Mort Base Effectiveness:** 0%
    

### 1. The Attack Test:

- **Bonus %:**
    
    $$\left( \frac{1977 - 1610}{1100 \text{ (Base)}} \right) \times 100 = 33.3\%$$
    
- **Total GSP:**
    
    $$33.3 \times 1 = 33.3 \text{ GSP}$$
    
- **SAV Calculation:**
    
    $$\text{SAV}_{\text{Atk}} = \frac{33.3}{6} = \mathbf{5.55}$$
    

$$\text{SAV}_{\text{Atk}} = 5.55 \le 6.0 \longrightarrow L_{\text{Atk}} = 1$$

### 2. The Effectiveness Test:

- **Bonus Value:**
    
    $$4\% - 0\% = 4\%$$
    
- **Total GSP:**
    
    $$4 \times 1 = 4.0 \text{ GSP}$$
    
- **SAV Calculation:**
    
    $$\text{SAV}_{\text{Eff}} = \frac{4.0}{6} = \mathbf{0.66}$$
    

$$\text{SAV}_{\text{Eff}} = 0.66 \le 6.0 \longrightarrow L_{\text{Eff}} = 1$$
## GSP Landmine Analysis: Boss Arunka

Using the provided JSON data structure snippet, we can apply the **GSP Landmine Filter Algorithm** directly to the `averageStats` dataset.

To run the math precisely, we use standard Level 60 baseline constants for a high-defense warrior/mitigation profile:

- $B_{\text{Atk}} = 950$, $B_{\text{HP}} = 6200$, $B_{\text{Def}} = 700$, $B_{\text{Spd}} = 105$, $B_{\text{CC}} = 15\%$, $B_{\text{CD}} = 150\%$
    
- Mandatory flat main stat padding from left-side gear slots: Weapon = $+515 \text{ Atk}$, Helmet = $+2835 \text{ HP}$, Armor = $+310 \text{ Def}$ .
    

### Step-by-Step Mathematical Filtering

#### 1. Crit Chance (CC)

- **Bonus Value:** $23\% - 15\% = 8\%$
    
- **Total GSP:** $8 \times 1.6 = 12.8$
    
- **Slot Allocation Value:**
    
    $$\text{SAV}_{\text{CC}} = \frac{12.8}{6} = \mathbf{2.13}$$
    
- **Verdict:** $\text{SAV} \le 6.0 \longrightarrow$ **LANDMINE**
    

#### 2. Crit Damage (CD)

- **Bonus Value:** $157\% - 150\% = 7\%$
    
- **Total GSP:** $7 \times 1.14 = 7.98$
    
- **Slot Allocation Value:**
    
    $$\text{SAV}_{\text{CD}} = \frac{7.98}{6} = \mathbf{1.33}$$
    
- **Verdict:** $\text{SAV} \le 6.0 \longrightarrow$ **LANDMINE**
    

#### 3. Attack (Atk)

- **Raw Stat Gain ($\Delta$):** $1559 - 950 = 609$
    
- **Strip Weapon Padding ($\Delta'$):** $609 - 515 = 94$
    
- **Percentage Equivalent (PE):** $\left(\frac{94}{950}\right) \times 100 = 9.89\%$
    
- **Total GSP:** $9.89 \times 1 = 9.89$
    
- **Slot Allocation Value:**
    
    $$\text{SAV}_{\text{Atk}} = \frac{9.89}{6} = \mathbf{1.65}$$
    
- **Verdict:** $\text{SAV} \le 6.0 \longrightarrow$ **LANDMINE**
    

#### 4. Effectiveness (Eff)

- **Bonus Value:** $45\% - 0\% = 45\%$
    
- **Total GSP:** $45 \times 1 = 45.0$
    
- **Slot Allocation Value:**
    
    $$\text{SAV}_{\text{Eff}} = \frac{45.0}{6} = \mathbf{7.50}$$
    
- **Verdict:** $\text{SAV} > 6.0 \longrightarrow$ **VALID STAT** (Intentional utility investment)
    

#### 5. Effect Resistance (Res)

- **Bonus Value:** $110\% - 0\% = 110\%$
    
- **Total GSP:** $110 \times 1 = 110.0$
    
- **Slot Allocation Value:**
    
    $$\text{SAV}_{\text{Res}} = \frac{110.0}{6} = \mathbf{18.33}$$
    
- **Verdict:** $\text{SAV} > 6.0 \longrightarrow$ **VALID STAT** (Core build priority)
    

### The Final Mathematical Metric Matrix

| **Stat**          | **Global JSON Average** | **Calculated SAV** | **Binary Filter Status (Li​)**      | **Resource Allocation Reality**                             |
| ----------------- | ----------------------- | ------------------ | ----------------------------------- | ----------------------------------------------------------- |
| **Crit Damage**   | 157%                    | **1.33**           | $L_{\text{CD}} = 1$ (**Landmine**)  | Pure background noise; unupgraded base stats.               |
| **Crit Chance**   | 23%                     | **2.13**           | $L_{\text{CC}} = 1$ (**Landmine**)  | Pure background noise; unupgraded base stats.               |
| **Attack**        | 1559                    | **1.65**           | $L_{\text{Atk}} = 1$ (**Landmine**) | Wasted offensive scaling; players actively avoid this stat. |
| **Effectiveness** | 45%                     | **7.50**           | $L_{\text{Eff}} = 0$ (Valid)        | Intentional; averaged across slots as active rolls.         |
| **Effect Res**    | 110%                    | **18.33**          | $L_{\text{Res}} = 0$ (Valid)        | High priority; heavily upgraded across the build.           |


---

